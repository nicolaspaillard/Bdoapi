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

var connectionString = fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable",
	os.Getenv("DB_USER"),
	os.Getenv("DB_PASSWORD"),
	os.Getenv("DB_HOST"),
	os.Getenv("DB_NAME"))

func main() {
	Migrate()
	Update()
	// Start()
}

func Migrate() {
	log.Println("Starting database migration process")
	db, err := sql.Open("postgres", connectionString)
	HandleError(ErrorParams{err: err, Message: "Failed to connect to database", fatal: true})
	defer db.Close()
	goose.SetDialect("postgres")
	err = goose.Up(db, "./migrations")
	HandleError(ErrorParams{err: err, Message: "Failed to run migrations", fatal: true})
	log.Println("Migration up completed successfully")
}
func Update() {
	log.Println("Starting updater")
	go func() {
		type PostItem struct {
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
		request := func(params RequestParams) {
			var resp *http.Response
			var err error
			switch params.Method {
			case "GET":
				resp, err = http.Get(params.Url)
			case "POST":
				resp, err = http.Post(params.Url, "application/json", bytes.NewBuffer(params.Body))
			}
			HandleError(ErrorParams{err: err, Message: "Failed to send request"})
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			HandleError(ErrorParams{err: err, Message: "Failed to read response body"})
			err = json.Unmarshal(body, &params.Output)
			HandleError(ErrorParams{err: err, Message: "Failed to unmarshal response body"})
		}
		for {
			var pearlItems1 WorldMarketListResult
			request(RequestParams{Method: "GET", Url: "https://api.arsha.io/v2/eu/GetWorldMarketList?lang=en&mainCategory=55&subCategory=1", Output: &pearlItems1})
			var pearlItems2 WorldMarketListResult
			request(RequestParams{Method: "GET", Url: "https://api.arsha.io/v2/eu/GetWorldMarketList?lang=en&mainCategory=55&subCategory=2", Output: &pearlItems2})
			var pearlItems = append(pearlItems1, pearlItems2...)
			x := 0
			for _, i := range pearlItems {
				if strings.Contains(i.Name, "Premium Set") {
					pearlItems[x] = i
					x++
				}
			}
			pearlItems = pearlItems[:x]
			var postItems = [][]PostItem{{}}
			for i, j := 0, 0; i < len(pearlItems); i++ {
				if len(postItems[j]) == 20 {
					postItems = append(postItems, []PostItem{})
					j++
				}
				postItems[j] = append(postItems[j], PostItem{Id: pearlItems[i].Id})
			}
			var preorderItems BiddingInfoListResult
			for _, postItem := range postItems {
				body, err := json.Marshal(postItem)
				HandleError(ErrorParams{err: err, Message: "Failed to marshal post body"})
				var output BiddingInfoListResult
				request(RequestParams{Method: "POST", Url: "https://api.arsha.io/v2/eu/GetBiddingInfoList?lang=en", Body: body, Output: &output})
				preorderItems = append(preorderItems, output...)
			}
			ctx := context.Background()
			conn, err := pgx.Connect(ctx, connectionString)
			HandleError(ErrorParams{err: err, Message: "Failed to connect to database"})
			queries := db.New(conn)
			currentTime := time.Now()
			scrapTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), currentTime.Hour(), 0, 0, 0, currentTime.Location())
			if len(pearlItems) == len(preorderItems) {
				for i, pearlItem := range pearlItems {
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
					HandleError(ErrorParams{err: err, Message: "Failed to create pearl item"})
				}
			}
			queries.DeleteOldPearlItems(ctx)
			conn.Close(ctx)
			time.Sleep(time.Hour)
		}
	}()
}
func Start() {
	getPearlItems := func(ginContext *gin.Context) {
		dbContext := context.Background()
		dbConn, err := pgx.Connect(dbContext, connectionString)
		HandleError(ErrorParams{err: err, Message: "Failed to connect to database"})
		defer dbConn.Close(dbContext)
		dbQueries := db.New(dbConn)
		date, err := time.Parse("2006-01-02 15:04:05", ginContext.Query("date"))
		log.Println(date)
		HandleError(ErrorParams{err: err, Message: "Failed to parse date"})
		pearlitems, err := dbQueries.GetPearlItems(dbContext, time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), 0, 0, 0, date.Location()))
		HandleError(ErrorParams{err: err, Message: "Failed to get pearl items"})
		ginContext.JSON(200, pearlitems)
	}
	gin.SetMode(gin.ReleaseMode)
	ginRouter := gin.Default()
	ginRouter.Use(cors.Default())
	ginRouter.GET("/pearlitems", getPearlItems)
	log.Printf("API started, listening on port %s", os.Getenv("PORT"))
	err := ginRouter.Run(fmt.Sprintf(":%s", os.Getenv("PORT")))
	HandleError(ErrorParams{err: err, Message: "Failed to start server", fatal: true})
}

type ErrorParams struct {
	err     error
	Message string
	fatal   bool
}

func HandleError(params ErrorParams) {
	if params.err != nil {
		if params.Message == "" {
			params.Message = "An error occured"
		}
		if params.fatal {
			log.Fatalf("%v %v", params.Message+":", params.err)
		} else {
			log.Printf("%v %v", params.Message+":", params.err)
		}
	}
}
