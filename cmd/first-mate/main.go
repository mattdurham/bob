package main

import (
	"context"
	"log"

	"github.com/mattdurham/bob/internal/firstmate/graph"
	"github.com/mattdurham/bob/internal/firstmate/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	dbPath, err := graph.DBPath()
	if err != nil {
		log.Fatalf("first-mate: determine db path: %v", err)
	}

	store, err := graph.Open(dbPath)
	if err != nil {
		log.Fatalf("first-mate: open store: %v", err)
	}
	defer store.Close()

	toolServer, err := tools.NewServer(store)
	if err != nil {
		log.Fatalf("first-mate: create tool server: %v", err)
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "first-mate",
		Version: "v0.1.0",
	}, nil)

	toolServer.Register(server)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("first-mate: server error: %v", err)
	}
}
