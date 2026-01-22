package service

import (
	"context"
	"encoding/json"
	"fmt"
	"iat/internal/model"
	"iat/internal/repo"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type MCPClientInterface interface {
	Start(ctx context.Context) error
	Close() error
	Initialize(ctx context.Context, request mcp.InitializeRequest) (*mcp.InitializeResult, error)
	ListTools(ctx context.Context, request mcp.ListToolsRequest) (*mcp.ListToolsResult, error)
	CallTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
}

type MCPService struct {
	repo      *repo.MCPServerRepo
	clients   map[uint]MCPClientInterface
	clientsMu sync.Mutex
}

func NewMCPService() *MCPService {
	return &MCPService{
		repo:    repo.NewMCPServerRepo(),
		clients: make(map[uint]MCPClientInterface),
	}
}

// --- CRUD Methods ---

func (s *MCPService) CreateMCPServer(name, description, serverType, command, args, env, url string) error {
	server := &model.MCPServer{
		Name:        name,
		Description: description,
		Type:        serverType,
		Command:     command,
		Args:        args,
		Env:         env,
		Url:         url,
		Enabled:     true,
	}
	return s.repo.Create(server)
}

func (s *MCPService) UpdateMCPServer(id uint, name, description, serverType, command, args, env, url string, enabled bool) error {
	server, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	// If configuration changes, disconnect existing client
	s.Disconnect(id)

	server.Name = name
	server.Description = description
	server.Type = serverType
	server.Command = command
	server.Args = args
	server.Env = env
	server.Url = url
	server.Enabled = enabled
	return s.repo.Update(server)
}

func (s *MCPService) DeleteMCPServer(id uint) error {
	s.Disconnect(id)
	return s.repo.Delete(id)
}

func (s *MCPService) ListMCPServers() ([]model.MCPServer, error) {
	return s.repo.List()
}

func (s *MCPService) GetMCPServer(id uint) (*model.MCPServer, error) {
	return s.repo.GetByID(id)
}

// --- Client Methods ---

func (s *MCPService) Disconnect(id uint) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	if cli, ok := s.clients[id]; ok {
		cli.Close()
		delete(s.clients, id)
	}
}

func (s *MCPService) getOrConnect(ctx context.Context, server *model.MCPServer) (MCPClientInterface, error) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	if cli, ok := s.clients[server.ID]; ok {
		// TODO: Check if alive?
		return cli, nil
	}

	var cli MCPClientInterface
	var err error

	if server.Type == "stdio" {
		var args []string
		if server.Args != "" {
			_ = json.Unmarshal([]byte(server.Args), &args)
		}
		var env map[string]string
		if server.Env != "" {
			_ = json.Unmarshal([]byte(server.Env), &env)
		}

		var envList []string
		for k, v := range env {
			envList = append(envList, fmt.Sprintf("%s=%s", k, v))
		}

		stdioCli, err := client.NewStdioMCPClient(server.Command, envList, args...)
		if err != nil {
			return nil, err
		}
		if err := stdioCli.Start(ctx); err != nil {
			return nil, fmt.Errorf("failed to start stdio client: %v", err)
		}
		cli = stdioCli
	} else if server.Type == "sse" {
		return nil, fmt.Errorf("SSE transport not implemented yet")
	} else {
		return nil, fmt.Errorf("unknown server type: %s", server.Type)
	}

	// Initialize
	initReq := mcp.InitializeRequest{
		Request: mcp.Request{
			Method: string(mcp.MethodInitialize),
		},
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "iat-client",
				Version: "1.0.0",
			},
			Capabilities: mcp.ClientCapabilities{
				Sampling: &struct{}{},
			},
		},
	}

	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err = cli.Initialize(initCtx, initReq)
	if err != nil {
		cli.Close()
		return nil, fmt.Errorf("failed to initialize mcp: %v", err)
	}

	s.clients[server.ID] = cli
	return cli, nil
}

func (s *MCPService) GetToolsForServers(servers []model.MCPServer) ([]*schema.ToolInfo, error) {
	var result []*schema.ToolInfo

	for _, server := range servers {
		if !server.Enabled {
			continue
		}

		cli, err := s.getOrConnect(context.Background(), &server)
		if err != nil {
			log.Printf("Failed to connect to MCP server %s: %v", server.Name, err)
			continue
		}

		// List Tools
		listResp, err := cli.ListTools(context.Background(), mcp.ListToolsRequest{
			PaginatedRequest: mcp.PaginatedRequest{
				Request: mcp.Request{
					Method: string(mcp.MethodToolsList),
				},
			},
		})
		if err != nil {
			log.Printf("Failed to list tools from %s: %v", server.Name, err)
			continue
		}

		for _, tool := range listResp.Tools {
			// Convert to Eino ToolInfo
			// We prefix name to route calls later: mcp__<serverID>__<toolName>
			uniqueName := fmt.Sprintf("mcp__%d__%s", server.ID, tool.Name)

			// Parse Schema
			// tool.InputSchema is ToolInputSchema which implements Marshaler
			schemaBytes, _ := json.Marshal(tool.InputSchema)
			var jsSchema jsonschema.Schema
			if err := json.Unmarshal(schemaBytes, &jsSchema); err != nil {
				log.Printf("Failed to parse schema for tool %s: %v", tool.Name, err)
				continue
			}

			result = append(result, &schema.ToolInfo{
				Name:        uniqueName,
				Desc:        fmt.Sprintf("[%s] %s", server.Name, tool.Description),
				ParamsOneOf: schema.NewParamsOneOfByJSONSchema(&jsSchema),
			})
		}
	}
	return result, nil
}

func (s *MCPService) GetGlobalTools(ctx context.Context) ([]*schema.ToolInfo, error) {
	servers, err := s.repo.ListEnabled()
	if err != nil {
		return nil, err
	}
	return s.GetToolsForServers(servers)
}

func (s *MCPService) ListToolsForServer(serverID uint) ([]*schema.ToolInfo, error) {
	server, err := s.repo.GetByID(serverID)
	if err != nil {
		return nil, err
	}
	if !server.Enabled {
		return nil, fmt.Errorf("server is disabled")
	}

	cli, err := s.getOrConnect(context.Background(), server)
	if err != nil {
		return nil, err
	}

	listResp, err := cli.ListTools(context.Background(), mcp.ListToolsRequest{
		PaginatedRequest: mcp.PaginatedRequest{
			Request: mcp.Request{
				Method: string(mcp.MethodToolsList),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	var result []*schema.ToolInfo
	for _, tool := range listResp.Tools {
		// No need to prefix for simple listing, but to be consistent with ChatService
		uniqueName := fmt.Sprintf("mcp(%d):%s", server.ID, tool.Name)

		schemaBytes, _ := json.Marshal(tool.InputSchema)
		var jsSchema jsonschema.Schema
		if err := json.Unmarshal(schemaBytes, &jsSchema); err != nil {
			log.Printf("Failed to parse schema for tool %s: %v", tool.Name, err)
			continue
		}

		result = append(result, &schema.ToolInfo{
			Name:        uniqueName,
			Desc:        tool.Description, // Use original description
			ParamsOneOf: schema.NewParamsOneOfByJSONSchema(&jsSchema),
		})
	}
	return result, nil
}

func (s *MCPService) CallTool(ctx context.Context, name string, args map[string]interface{}) (string, error) {
	// Name format: mcp__<serverID>__<toolName>
	parts := strings.SplitN(name, "__", 3)
	if len(parts) != 3 || parts[0] != "mcp" {
		return "", fmt.Errorf("invalid mcp tool name: %s", name)
	}

	var serverID uint
	_, err := fmt.Sscanf(parts[1], "%d", &serverID)
	if err != nil {
		return "", fmt.Errorf("invalid server id in tool name: %s", name)
	}
	realToolName := parts[2]

	s.clientsMu.Lock()
	cli, ok := s.clients[serverID]
	s.clientsMu.Unlock()

	if !ok {
		// Try to reconnect?
		// We need the server config.
		server, err := s.repo.GetByID(serverID)
		if err != nil {
			return "", fmt.Errorf("server not found: %v", err)
		}
		cli, err = s.getOrConnect(ctx, server)
		if err != nil {
			return "", fmt.Errorf("failed to connect to server: %v", err)
		}
	}

	callReq := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: string(mcp.MethodToolsCall),
		},
		Params: mcp.CallToolParams{
			Name:      realToolName,
			Arguments: args,
		},
	}

	res, err := cli.CallTool(ctx, callReq)
	if err != nil {
		return "", err
	}

	// Combine content
	var output strings.Builder
	for _, c := range res.Content {
		switch v := c.(type) {
		case mcp.TextContent:
			output.WriteString(v.Text)
		case mcp.ImageContent:
			output.WriteString(fmt.Sprintf("[Image: %s]", v.MIMEType))
		case mcp.EmbeddedResource:
			output.WriteString("[Embedded Resource]")
		default:
			output.WriteString(fmt.Sprintf("[Unknown Content: %T]", v))
		}
	}

	if res.IsError {
		return "", fmt.Errorf("tool execution failed: %s", output.String())
	}

	return output.String(), nil
}
