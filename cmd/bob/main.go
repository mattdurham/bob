package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/server"
)

const version = "0.1.0"

var (
	serve       = flag.Bool("serve", false, "Run as MCP server (stdio mode)")
	showVersion = flag.Bool("version", false, "Show version")
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("bob v%s\n", version)
		return
	}

	// MCP server mode (stdio) - writes to shared database
	if *serve {
		// Open database connection
		db, err := NewDatabase()
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		defer db.Close()

		// Create MCP server with database
		s := CreateMCPServerWithDB(db)
		if err := server.ServeStdio(s); err != nil {
			log.Fatalf("MCP server error: %v", err)
		}
		return
	}

	// Default: show usage
	fmt.Printf("🏴‍☠️ bob v%s - Belayin' Pin Bob, Captain of Your Agents\n\n", version)
	fmt.Printf("Usage:\n")
	fmt.Printf("  bob --serve           Run as MCP server (for Claude integration)\n")
	fmt.Printf("  bob --version         Show version\n\n")
	fmt.Printf("Architecture:\n")
	fmt.Printf("  • Each Claude session runs 'bob --serve' (MCP stdio mode)\n")
	fmt.Printf("  • All sessions write to ~/.bob/state/db.sql (shared SQLite)\n")
	fmt.Printf("  • Bob orchestrates workflows and manages tasks across sessions\n")
}
