package main

import (
	"bdoapi/db"
	"bdoapi/internal/config"
	"bdoapi/internal/events"
	"bdoapi/internal/lightstones"
	"bdoapi/internal/pearl"
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func Migrate() {
	log.Println("Starting database migration process")
	db, err := sql.Open("postgres", config.ConnectionString)
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
	updaters := []struct {
		name     string
		run      func() error
		interval time.Duration
		aligned  bool
		retry    time.Duration
	}{
		{"outfits", pearl.Outfits.Update, time.Hour, true, 5 * time.Minute},
		{"pets", pearl.Pets.Update, time.Hour, true, 5 * time.Minute},
		{"events", events.Update, 24 * time.Hour, true, 30 * time.Minute},
		{"lightstones", lightstones.Update, 5 * time.Minute, false, 5 * time.Minute},
	}
	for _, u := range updaters {
		go func() {
			for {
				if err := u.run(); err != nil {
					log.Printf("Update %s failed: %v", u.name, err)
					time.Sleep(u.retry)
					continue
				}
				log.Printf("Updated %s successfully", u.name)
				if u.aligned {
					time.Sleep(time.Until(time.Now().Truncate(u.interval).Add(u.interval)))
				} else {
					time.Sleep(u.interval)
				}
			}
		}()
	}
}

func Start() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(cors.Default())
	router.GET("/outfits", pearl.Handler("outfits", func(ctx context.Context, q *db.Queries, date time.Time) (any, error) {
		return q.GetOutfits(ctx, date)
	}))
	router.GET("/pets", pearl.Handler("pets", func(ctx context.Context, q *db.Queries, date time.Time) (any, error) {
		return q.GetPets(ctx, date)
	}))
	router.GET("/events", events.Handler)
	router.GET("/lightstones", lightstones.Handler)
	port := config.GetEnv("PORT", "8081")
	log.Printf("API started, listening on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Printf("Failed to start server: %v", err)
	}
}

func main() {
	Migrate()
	Update()
	Start()
}
