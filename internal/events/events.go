package events

import (
	"bdoapi/db"
	"bdoapi/internal/config"
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/jackc/pgx/v5"
)

const (
	eventsURL      = "https://www.naeu.playblackdesert.com/en-US/News/Notice?boardType=3&Page=%d"
	eventSelector  = `a[name="btnDetail"]`
	maxEventPages  = 2
	scrapeDeadline = 2 * time.Minute
)

type scrapedEvent struct {
	ID    int64
	URL   string
	Image string
	Tag   string
	Title string
	End   time.Time
}

func text(el *rod.Element, sel string) string {
	if ok, found, _ := el.Has(sel); ok {
		return strings.TrimSpace(found.MustText())
	}
	return ""
}

func attr(el *rod.Element, sel, name string) string {
	target := el
	if sel != "" {
		ok, found, _ := el.Has(sel)
		if !ok {
			return ""
		}
		target = found
	}
	if v := target.MustAttribute(name); v != nil {
		return *v
	}
	return ""
}

// parseEvents reads every event with a "N days left" counter on the page.
// rod.Try turns Must* panics into errors so a bad page can't crash the server.
func parseEvents(page *rod.Page, today time.Time) (events []scrapedEvent, err error) {
	err = rod.Try(func() {
		for _, a := range page.MustElements(eventSelector) {
			days, convErr := strconv.Atoi(text(a, ".count em"))
			if convErr != nil {
				continue // "Ongoing" or no counter
			}
			id, idErr := strconv.ParseInt(attr(a, "", "data-boardno"), 10, 64)
			if idErr != nil {
				continue
			}
			events = append(events, scrapedEvent{
				ID:    id,
				URL:   attr(a, "", "href"),
				Image: attr(a, "img", "src"),
				Tag:   text(a, `i[class^="event_tag"]`),
				Title: text(a, ".title em"),
				End:   today.AddDate(0, 0, days),
			})
		}
	})
	return events, err
}

func scrapeEvents() ([]scrapedEvent, error) {
	l := launcher.New().Leakless(true).Headless(true)
	if path, ok := launcher.LookPath(); ok {
		l = l.Bin(path)
	}
	defer l.Cleanup()

	u, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("launch browser: %w", err)
	}
	browser := rod.New().ControlURL(u).Timeout(scrapeDeadline)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("connect to browser: %w", err)
	}
	defer browser.Close()

	today := time.Now().UTC().Truncate(24 * time.Hour)
	seen := map[int64]bool{}
	var all []scrapedEvent

	for p := 1; p <= maxEventPages; p++ {
		page, err := browser.Page(proto.TargetCreateTarget{URL: fmt.Sprintf(eventsURL, p)})
		if err != nil {
			return nil, fmt.Errorf("open page %d: %w", p, err)
		}
		if err := page.WaitLoad(); err != nil {
			page.Close()
			return nil, fmt.Errorf("load page %d: %w", p, err)
		}
		if _, err := page.Timeout(10 * time.Second).Element(eventSelector); err != nil {
			page.Close()
			if p == 1 {
				return nil, fmt.Errorf("no events found on first page: %w", err)
			}
			break // ran past the last page
		}

		events, err := parseEvents(page, today)
		page.Close()
		if err != nil {
			return nil, fmt.Errorf("parse page %d: %w", p, err)
		}

		added := 0
		for _, e := range events {
			if !seen[e.ID] {
				seen[e.ID] = true
				all = append(all, e)
				added++
			}
		}
		if added == 0 {
			break // page had nothing new
		}
	}
	return all, nil
}

// Update scrapes the events pages and stores them in the database.
func Update() error {
	events, err := scrapeEvents()
	if err != nil {
		return err
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, config.ConnectionString)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer conn.Close(ctx)
	queries := db.New(conn)

	failed := 0
	for _, e := range events {
		if err := queries.CreateEvent(ctx, db.CreateEventParams{
			Itemid: e.ID,
			Url:    e.URL,
			Image:  e.Image,
			Tag:    e.Tag,
			Title:  e.Title,
			End:    e.End,
		}); err != nil {
			log.Printf("Failed to create event %d: %v", e.ID, err)
			failed++
		}
	}
	if err := queries.DeleteExpiredEvents(ctx); err != nil {
		return fmt.Errorf("delete expired events: %w", err)
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d events failed to save", failed, len(events))
	}
	return nil
}

// Handler serves the stored events.
func Handler(c *gin.Context) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, config.ConnectionString)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	defer conn.Close(ctx)

	events, err := db.New(conn).GetEvents(ctx)
	if err != nil {
		log.Printf("Failed to get events: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, events)
}
