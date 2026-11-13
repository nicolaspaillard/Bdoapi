package lightstones

import (
	"bdoapi/db"
	"bdoapi/internal/arsha"
	"bdoapi/internal/config"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// Update replaces the stored lightstones with the current market values.
// Lightstones missing from the market list no longer exist in the game and are deleted.
func Update() error {
	items, err := arsha.GetWorldMarketList(85, 1, 2, 3, 4)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return fmt.Errorf("empty market list")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, config.ConnectionString)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer conn.Close(ctx)
	queries := db.New(conn)

	ids := make([]int64, 0, len(items))
	failed := 0
	for _, m := range items {
		if strings.HasPrefix(m.Name, "Purified") {
			continue
		}
		ids = append(ids, m.Id)
		if err := queries.UpsertLightstone(ctx, db.UpsertLightstoneParams{
			Itemid: m.Id,
			Name:   m.Name,
			Stock:  m.CurrentStock,
			Price:  m.BasePrice,
		}); err != nil {
			log.Printf("Failed to save lightstone %d: %v", m.Id, err)
			failed++
		}
	}
	if err := queries.DeleteMissingLightstones(ctx, ids); err != nil {
		return fmt.Errorf("delete missing lightstones: %w", err)
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d lightstones failed to save", failed, len(items))
	}
	return nil
}

// Handler serves the stored lightstones.
func Handler(c *gin.Context) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, config.ConnectionString)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	defer conn.Close(ctx)

	items, err := db.New(conn).GetLightstones(ctx)
	if err != nil {
		log.Printf("Failed to get lightstones: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, items)
}
