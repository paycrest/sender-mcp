// Command paycrest-mcp is the Paycrest Model Context Protocol server (stdio).
// It exposes selected Paycrest aggregator v2 HTTP endpoints as MCP tools.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/joho/godotenv"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paycrest/paycrest/sender-mcp/config"
	"github.com/paycrest/paycrest/sender-mcp/mcpserver"
	"github.com/paycrest/paycrest/sender-mcp/paycrest"
)

func main() {
	// Optional: load `.env` from current working directory (ignored if missing).
	_ = godotenv.Load()

	cfg := config.Load()
	client := paycrest.NewClient(cfg)

	s := mcp.NewServer(&mcp.Implementation{
		Name:    "paycrest",
		Version: "0.1.0",
	}, nil)
	mcpserver.RegisterTools(s, client)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := s.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("mcp server: %v", err)
	}
}
