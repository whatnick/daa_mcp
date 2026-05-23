package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/whatnick/daa_mcp/internal/atlasclient"
	"github.com/whatnick/daa_mcp/internal/mcpserver"
	"github.com/whatnick/daa_mcp/internal/model"
)

func main() {
	baseURL := os.Getenv("DIGITAL_ATLAS_BASE_URL")
	collection := flag.String("collection", "dataset", "Collection ID (dataset, appAndMap, document, all)")
	query := flag.String("q", "", "Free-text search query")
	limit := flag.Int("limit", 10, "Page size")
	startIndex := flag.Int("startindex", 1, "Pagination start index")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client := atlasclient.New(baseURL)
	server := mcpserver.New(client)

	emit := func(event model.StreamEvent) error {
		b, err := json.Marshal(event)
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	}

	if err := server.StreamSearchItems(ctx, *collection, *query, *limit, *startIndex, emit); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
