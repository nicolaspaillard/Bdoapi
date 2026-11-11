package main

import (
	"bdoapi/db"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

var connectionString = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
	getEnv("DB_USER", "postgres"),
	getEnv("DB_PASSWORD", "password"),
	getEnv("DB_HOST", "localhost"),
	getEnv("DB_PORT", "5432"),
	getEnv("DB_NAME", "bdoapi"))

func main() {
	Migrate()
	Update()
	Start()
}

func Migrate() {
	log.Println("Starting database migration process")
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return
	}
	defer db.Close()
	goose.SetDialect("postgres")
	err = goose.Up(db, "./sqlc/migrations")
	if err != nil {
		log.Printf("Failed to run migrations: %v", err)
		return
	}
	log.Println("Migration up completed successfully")
}
func Update() {
	log.Println("Starting updater")
	go func() {
		type PostId struct {
			Id int64 `json:"id"`
		}
		type RequestParams struct {
			Method string
			Url    string
			Body   []byte
			Output interface{}
		}
		type BiddingInfoListResult []struct {
			Name   string `json:"name"`
			Id     int64  `json:"id"`
			Sid    int64  `json:"sid"`
			Orders []struct {
				Price   int64 `json:"price"`
				Sellers int64 `json:"sellers"`
				Buyers  int64 `json:"buyers"`
			} `json:"orders"`
		}
		type WorldMarketListResult []struct {
			Name         string `json:"name"`
			Id           int64  `json:"id"`
			CurrentStock int64  `json:"currentStock"`
			TotalTrades  int64  `json:"totalTrades"`
			BasePrice    int64  `json:"basePrice"`
			MainCategory int    `json:"mainCategory"`
			SubCategory  int    `json:"subCategory"`
		}
		request := func(params RequestParams) error {
			var resp *http.Response
			var err error
			switch params.Method {
			case "GET":
				resp, err = http.Get(params.Url)
			case "POST":
				resp, err = http.Post(params.Url, "application/json", bytes.NewBuffer(params.Body))
			}
			if err != nil {
				return fmt.Errorf("failed to send request: %w", err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body: %w", err)
			}
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
			}
			err = json.Unmarshal(body, params.Output)
			if err != nil {
				return fmt.Errorf("failed to unmarshal response body: %w. %s", err, string(body))
			}
			return nil
		}
		for {
			var outfitsMale WorldMarketListResult
			if err := request(RequestParams{Method: "GET", Url: "https://api.arsha.io/v2/eu/GetWorldMarketList?lang=en&mainCategory=55&subCategory=1", Output: &outfitsMale}); err != nil {
				log.Printf("Failed to get world market list (subCategory=1): %v", err)
				time.Sleep(5 * time.Minute)
				continue
			}
			var outfitsFemale WorldMarketListResult
			if err := request(RequestParams{Method: "GET", Url: "https://api.arsha.io/v2/eu/GetWorldMarketList?lang=en&mainCategory=55&subCategory=2", Output: &outfitsFemale}); err != nil {
				log.Printf("Failed to get world market list (subCategory=2): %v", err)
				time.Sleep(5 * time.Minute)
				continue
			}
			var outfits = append(outfitsMale, outfitsFemale...)
			i := 0
			for _, outfit := range outfits {
				if strings.Contains(outfit.Name, "Premium Set") {
					outfits[i] = outfit
					i++
				}
			}
			outfits = outfits[:i]
			var postIds = [][]PostId{{}}
			for i, j := 0, 0; i < len(outfits); i++ {
				if i > 0 && i%20 == 0 {
					postIds = append(postIds, []PostId{})
					j++
				}
				postIds[j] = append(postIds[j], PostId{Id: outfits[i].Id})
			}
			var preorderItems BiddingInfoListResult
			failed := false
			for _, postIds := range postIds {
				body, err := json.Marshal(postIds)
				if err != nil {
					log.Printf("Failed to marshal post body: %v", err)
					continue
				}
				var output BiddingInfoListResult
				if err := request(RequestParams{Method: "POST", Url: "https://api.arsha.io/v2/eu/GetBiddingInfoList?lang=en", Body: body, Output: &output}); err != nil {
					log.Printf("Failed to get bidding info list: %v", err)
					failed = true
					break
				}
				preorderItems = append(preorderItems, output...)
			}
			if failed {
				time.Sleep(5 * time.Minute)
				continue
			}
			ctx := context.Background()
			conn, err := pgx.Connect(ctx, connectionString)
			if err != nil {
				log.Printf("Failed to connect to database: %v", err)
				continue
			}
			queries := db.New(conn)
			currentTime := time.Now()
			scrapTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), currentTime.Hour(), 0, 0, 0, currentTime.Location())
			if len(outfits) == len(preorderItems) {
				for i, pearlItem := range outfits {
					var preorders int64
					for _, j := range preorderItems[i].Orders {
						preorders += j.Buyers
					}
					err := queries.CreatePearlItem(ctx, db.CreatePearlItemParams{
						Itemid:    pearlItem.Id,
						Name:      pearlItem.Name,
						Date:      scrapTime,
						Sold:      pearlItem.TotalTrades,
						Preorders: preorders,
					})
					if err != nil {
						log.Printf("Failed to create pearl item: %v", err)
						continue
					}
				}
			}
			queries.DeleteOldPearlItems(ctx)
			conn.Close(ctx)
			log.Println("Pearl items updated successfully")
			log.Println("Sleeping until the next hour")
			time.Sleep(time.Until(time.Now().Truncate(time.Hour).Add(time.Hour)))
		}
	}()
}
func Start() {
	getPearlItems := func(ginContext *gin.Context) {
		dbContext := context.Background()
		dbConn, err := pgx.Connect(dbContext, connectionString)
		if err != nil {
			log.Printf("Failed to connect to database: %v", err)
			return
		}
		defer dbConn.Close(dbContext)
		dbQueries := db.New(dbConn)
		date, err := time.Parse("2006-01-02 15:04:05", ginContext.Query("date"))
		log.Println(date)
		if err != nil {
			log.Printf("Failed to parse date: %v", err)
			return
		}
		pearlitems, err := dbQueries.GetPearlItems(dbContext, time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), 0, 0, 0, date.Location()))
		if err != nil {
			log.Printf("Failed to get pearl items: %v", err)
			return
		}
		ginContext.JSON(200, pearlitems)
	}
	gin.SetMode(gin.ReleaseMode)
	ginRouter := gin.Default()
	ginRouter.Use(cors.Default())
	ginRouter.GET("/pearlitems", getPearlItems)
	err := ginRouter.Run(fmt.Sprintf(":%s", getEnv("PORT", "8081")))
	if err != nil {
		log.Printf("Failed to start server: %v", err)
		return
	}
	log.Printf("API started, listening on port %s", getEnv("PORT", "8081"))
}
