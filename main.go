package main

import (
	"bdoapi/db"
	"bdoapi/internal/config"
	"bdoapi/internal/events"
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
	go func() {
		updaters := []struct {
			name string
			run  func() error
		}{
			{"outfits", pearl.UpdateOutfits},
			{"pets", pearl.UpdatePets},
			{"events", events.Update},
		}
		for {
			failed := false
			for _, u := range updaters {
				if err := u.run(); err != nil {
					log.Printf("Update %s failed: %v", u.name, err)
					failed = true
					continue
				}
				log.Printf("Updated %s successfully", u.name)
			}
			if failed {
				time.Sleep(5 * time.Minute)
				continue
			}
			time.Sleep(time.Until(time.Now().Truncate(time.Hour).Add(time.Hour)))
		}
	}()
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
