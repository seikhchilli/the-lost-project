package main

import (
	"context"
	"log"
	"titles-mcp/config"
	"titles-mcp/database"
	"titles-mcp/tools"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	config.LoadConfig()

	db := database.NewDb()

	server := mcp.NewServer(&mcp.Implementation{Name: "title-mcp", Version: "v1.0.0"}, nil)

	titleTool := tools.NewTitleTool(db)
	titleTool.Register(server)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
