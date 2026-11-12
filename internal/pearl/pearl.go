package pearl

import (
	"bdoapi/db"
	"bdoapi/internal/config"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type PostId struct {
	Id int64 `json:"id"`
}

type BiddingInfo struct {
	Name   string `json:"name"`
	Id     int64  `json:"id"`
	Sid    int64  `json:"sid"`
	Orders []struct {
		Price   int64 `json:"price"`
		Sellers int64 `json:"sellers"`
		Buyers  int64 `json:"buyers"`
	} `json:"orders"`
}

type MarketItem struct {
	Name         string `json:"name"`
	Id           int64  `json:"id"`
	CurrentStock int64  `json:"currentStock"`
	TotalTrades  int64  `json:"totalTrades"`
	BasePrice    int64  `json:"basePrice"`
	MainCategory int    `json:"mainCategory"`
	SubCategory  int    `json:"subCategory"`
}

const apiBase = "https://api.arsha.io/v2/eu"

var httpClient = &http.Client{Timeout: 30 * time.Second}

func request(method, url string, body []byte, out any) error {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(data))
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("failed to unmarshal response body: %w. %s", err, string(data))
	}
	return nil
}

func getWorldMarketList(mainCategory int, subCategories ...int) ([]MarketItem, error) {
	fetch := func(url string) ([]MarketItem, error) {
		var items []MarketItem
		if err := request(http.MethodGet, url, nil, &items); err != nil {
			return nil, fmt.Errorf("world market list (%s): %w", url, err)
		}
		return items, nil
	}

	base := fmt.Sprintf("%s/GetWorldMarketList?lang=en&mainCategory=%d", apiBase, mainCategory)
	if len(subCategories) == 0 {
		return fetch(base)
	}

	var all []MarketItem
	for _, sub := range subCategories {
		items, err := fetch(fmt.Sprintf("%s&subCategory=%d", base, sub))
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
	}
	return all, nil
}

func getBiddingInfo(items []MarketItem) ([]BiddingInfo, error) {
	var result []BiddingInfo
	for start := 0; start < len(items); start += 20 {
		end := min(start+20, len(items))
		ids := make([]PostId, 0, end-start)
		for _, item := range items[start:end] {
			ids = append(ids, PostId{Id: item.Id})
		}
		body, err := json.Marshal(ids)
		if err != nil {
			return nil, fmt.Errorf("marshal post body: %w", err)
		}
		var out []BiddingInfo
		if err := request(http.MethodPost, apiBase+"/GetBiddingInfoList?lang=en", body, &out); err != nil {
			return nil, fmt.Errorf("bidding info list: %w", err)
		}
		result = append(result, out...)
	}
	return result, nil
}

func hourOf(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
}

// Shared update logic: fetches bidding info, then saves every item through the given callbacks.
func updateItems(label string, items []MarketItem, save func(ctx context.Context, q *db.Queries, item MarketItem, preorders int64, date time.Time) error, cleanup func(ctx context.Context, q *db.Queries)) error {
	bids, err := getBiddingInfo(items)
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
	for i, item := range items {
		var preorders int64
		for _, order := range bids[i].Orders {
			preorders += order.Buyers
		}
		if err := save(ctx, queries, item, preorders, scrapTime); err != nil {
			log.Printf("Failed to create %s: %v", label, err)
		}
	}
	cleanup(ctx, queries)
	return nil
}

func UpdateOutfits() error {
	market, err := getWorldMarketList(55, 1, 2)
	if err != nil {
		return err
	}
	var outfits []MarketItem
	for _, item := range market {
		if strings.Contains(item.Name, "Premium Set") {
			outfits = append(outfits, item)
		}
	}
	return updateItems("outfits", outfits,
		func(ctx context.Context, q *db.Queries, item MarketItem, preorders int64, date time.Time) error {
			return q.CreateOutfit(ctx, db.CreateOutfitParams{
				Itemid:    item.Id,
				Name:      item.Name,
				Date:      date,
				Sold:      item.TotalTrades,
				Preorders: preorders,
			})
		},
		func(ctx context.Context, q *db.Queries) { q.DeleteOldOutfits(ctx) },
	)
}

func UpdatePets() error {
	pets, err := getWorldMarketList(55, 8)
	if err != nil {
		return err
	}
	return updateItems("pets", pets,
		func(ctx context.Context, q *db.Queries, item MarketItem, preorders int64, date time.Time) error {
			return q.CreatePet(ctx, db.CreatePetParams{
				Itemid:    item.Id,
				Name:      item.Name,
				Date:      date,
				Sold:      item.TotalTrades,
				Preorders: preorders,
			})
		},
		func(ctx context.Context, q *db.Queries) { q.DeleteOldPets(ctx) },
	)
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
