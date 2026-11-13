package pearl

import (
	"bdoapi/db"
	"bdoapi/internal/arsha"
	"bdoapi/internal/config"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func hourOf(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
}

// Shared update logic: fetches bidding info, then saves every item through the given callbacks.
func updateItems(label string, items []arsha.MarketItem, save func(ctx context.Context, q *db.Queries, item arsha.MarketItem, preorders int64, date time.Time) error, cleanup func(ctx context.Context, q *db.Queries) error) error {
	bids, err := arsha.GetBiddingInfo(items)
	if err != nil {
		return err
	}
	if len(items) != len(bids) {
		return fmt.Errorf("mismatch: %d %s, %d bidding infos", len(items), label, len(bids))
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, config.ConnectionString)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer conn.Close(ctx)
	queries := db.New(conn)

	scrapTime := hourOf(time.Now())
	failed := 0
	for i, item := range items {
		var preorders int64
		for _, order := range bids[i].Orders {
			preorders += order.Buyers
		}
		if err := save(ctx, queries, item, preorders, scrapTime); err != nil {
			log.Printf("Failed to create %s: %v", label, err)
			failed++
		}
	}
	if err := cleanup(ctx, queries); err != nil {
		return fmt.Errorf("cleanup %s: %w", label, err)
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d %s failed to save", failed, len(items), label)
	}
	return nil
}

type item struct {
	label   string
	subs    []int
	filter  func(arsha.MarketItem) bool
	save    func(ctx context.Context, q *db.Queries, item arsha.MarketItem, preorders int64, date time.Time) error
	cleanup func(ctx context.Context, q *db.Queries) error
}

var Outfits = item{
	label:  "outfits",
	subs:   []int{1, 2},
	filter: func(i arsha.MarketItem) bool { return strings.Contains(i.Name, "Premium Set") },
	save: func(ctx context.Context, q *db.Queries, m arsha.MarketItem, preorders int64, date time.Time) error {
		return q.CreateOutfit(ctx, db.CreateOutfitParams{
			Itemid:    m.Id,
			Name:      m.Name,
			Date:      date,
			Sold:      m.TotalTrades,
			Preorders: preorders,
		})
	},
	cleanup: func(ctx context.Context, q *db.Queries) error { return q.DeleteOldOutfits(ctx) },
}

var Pets = item{
	label: "pets",
	subs:  []int{8},
	save: func(ctx context.Context, q *db.Queries, m arsha.MarketItem, preorders int64, date time.Time) error {
		return q.CreatePet(ctx, db.CreatePetParams{
			Itemid:    m.Id,
			Name:      m.Name,
			Date:      date,
			Sold:      m.TotalTrades,
			Preorders: preorders,
		})
	},
	cleanup: func(ctx context.Context, q *db.Queries) error { return q.DeleteOldPets(ctx) },
}

func (it item) Update() error {
	market, err := arsha.GetWorldMarketList(55, it.subs...)
	if err != nil {
		return err
	}
	items := market
	if it.filter != nil {
		items = nil
		for _, m := range market {
			if it.filter(m) {
				items = append(items, m)
			}
		}
	}
	return updateItems(it.label, items, it.save, it.cleanup)
}

// Handler serves the items returned by get for the date given in the "date" query parameter.
func Handler(label string, get func(ctx context.Context, q *db.Queries, date time.Time) (any, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		date, err := time.Parse("2006-01-02 15:04:05", c.Query("date"))
		if err != nil {
			log.Printf("Failed to parse date: %v", err)
			c.Status(http.StatusBadRequest)
			return
		}

		ctx := context.Background()
		conn, err := pgx.Connect(ctx, config.ConnectionString)
		if err != nil {
			log.Printf("Failed to connect to database: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		defer conn.Close(ctx)

		items, err := get(ctx, db.New(conn), hourOf(date))
		if err != nil {
			log.Printf("Failed to get %s: %v", label, err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, items)
	}
}
