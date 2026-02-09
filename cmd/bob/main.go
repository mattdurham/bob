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
	ui          = flag.Bool("ui", false, "Start web UI server")
	uiPort      = flag.String("port", "8080", "Web UI port")
	uiHost      = flag.String("host", "127.0.0.1", "Web UI host")
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("bob v%s\n", version)
		return
	}

	// Web UI mode
	if *ui {
		if err := StartUIServer(*uiHost, *uiPort); err != nil {
			log.Fatalf("UI server error: %v", err)
		}
		return
	}

	// MCP server mode (stdio)
	if *serve {
		// Create MCP server
		s := CreateMCPServer()
		if err := server.ServeStdio(s); err != nil {
			log.Fatalf("MCP server error: %v", err)
		}
		return
	}

	// Default: show usage
	fmt.Printf("🏴‍☠️ bob v%s - Belayin' Pin Bob, Captain of Your Agents\n\n", version)
	fmt.Printf("Usage:\n")
	fmt.Printf("  bob --serve           Run as MCP server (for Claude integration)\n")
	fmt.Printf("  bob --ui              Start web UI server (default: http://127.0.0.1:8080)\n")
	fmt.Printf("  bob --version         Show version\n\n")
	fmt.Printf("Web UI Options:\n")
	fmt.Printf("  --port PORT           Web UI port (default: 8080)\n")
	fmt.Printf("  --host HOST           Web UI host (default: 127.0.0.1)\n\n")
	fmt.Printf("Architecture:\n")
	fmt.Printf("  • Each Claude session runs 'bob --serve' (MCP stdio mode)\n")
	fmt.Printf("  • All sessions write to ~/.bob/state/ (JSON state files)\n")
	fmt.Printf("  • Bob orchestrates workflows and manages tasks across sessions\n")
	fmt.Printf("  • Use 'bob --ui' to view workflows and tasks in your browser\n")
}
