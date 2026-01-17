package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// 1. Create a new MCP server
	s := server.NewMCPServer(
		"test-mcp-server",
		"1.0.0",
		server.WithLogging(),
	)

	// 2. Define tools

	// Tool: hello
	helloTool := mcp.NewTool("hello",
		mcp.WithDescription("Greet the user"),
		mcp.WithString("name",
			mcp.Description("The name of the person to greet"),
			mcp.Required(),
		),
	)

	s.AddTool(helloTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := request.GetString("name", "Guest")
		return mcp.NewToolResultText(fmt.Sprintf("Hello, %s! Welcome to the MCP test server.", name)), nil
	})

	// Tool: add
	addTool := mcp.NewTool("add",
		mcp.WithDescription("Add two numbers"),
		mcp.WithNumber("a",
			mcp.Description("First number"),
			mcp.Required(),
		),
		mcp.WithNumber("b",
			mcp.Description("Second number"),
			mcp.Required(),
		),
	)

	s.AddTool(addTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		a := request.GetFloat("a", 0)
		b := request.GetFloat("b", 0)
		sum := a + b
		return mcp.NewToolResultText(fmt.Sprintf("%.2f", sum)), nil
	})

	// Tool: get_system_info
	sysInfoTool := mcp.NewTool("get_system_info",
		mcp.WithDescription("Get mock system information"),
	)

	s.AddTool(sysInfoTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		info := `
System Info:
- OS: Windows 11 Pro
- CPU: Intel Core i9-13900K
- RAM: 64GB
- Uptime: 24h 12m
`
		return mcp.NewToolResultText(info), nil
	})

	// 3. Start the server
	log.Println("Starting Test MCP Server...")
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
