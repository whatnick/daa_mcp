package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/whatnick/daa_mcp/internal/atlasclient"
	"github.com/whatnick/daa_mcp/internal/mcpserver"
)

var version = "dev"

type rpcRequest struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method"`
	Params  json.RawMessage  `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type toolsCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	baseURL := os.Getenv("DIGITAL_ATLAS_BASE_URL")
	client := atlasclient.New(baseURL)
	server := mcpserver.New(client)

	if err := serve(ctx, os.Stdin, os.Stdout, server); err != nil && !errors.Is(err, io.EOF) {
		fmt.Fprintf(os.Stderr, "mcp server error: %v\n", err)
		os.Exit(1)
	}
}

func serve(ctx context.Context, in io.Reader, out io.Writer, server *mcpserver.Server) error {
	reader := bufio.NewReader(in)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		msg, err := readMessage(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		var req rpcRequest
		if err := json.Unmarshal(msg, &req); err != nil {
			continue
		}
		if req.Method == "" {
			continue
		}

		// Notifications have no response id.
		if req.ID == nil {
			continue
		}

		idVal, err := parseID(*req.ID)
		if err != nil {
			if werr := writeResponse(out, rpcResponse{
				JSONRPC: "2.0",
				ID:      nil,
				Error:   &rpcError{Code: -32600, Message: "invalid request id"},
			}); werr != nil {
				return werr
			}
			continue
		}

		resp := handleRequest(ctx, req, idVal, server)
		if err := writeResponse(out, resp); err != nil {
			return err
		}
	}
}

func handleRequest(ctx context.Context, req rpcRequest, id interface{}, server *mcpserver.Server) rpcResponse {
	switch req.Method {
	case "initialize":
		return rpcResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{
						"listChanged": false,
					},
				},
				"serverInfo": map[string]interface{}{
					"name":    "daa_mcp",
					"version": version,
				},
			},
		}
	case "tools/list":
		return rpcResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: map[string]interface{}{
				"tools": toolDefinitions(),
			},
		}
	case "tools/call":
		result, err := callTool(ctx, req.Params, server)
		if err != nil {
			return rpcResponse{
				JSONRPC: "2.0",
				ID:      id,
				Result: map[string]interface{}{
					"isError": true,
					"content": []map[string]string{
						{"type": "text", "text": err.Error()},
					},
				},
			}
		}
		return rpcResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: map[string]interface{}{
				"content": []map[string]string{
					{"type": "text", "text": result},
				},
			},
		}
	default:
		return rpcResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error:   &rpcError{Code: -32601, Message: "method not found"},
		}
	}
}

func callTool(ctx context.Context, raw json.RawMessage, server *mcpserver.Server) (string, error) {
	var req toolsCallParams
	if err := json.Unmarshal(raw, &req); err != nil {
		return "", fmt.Errorf("invalid tool call parameters")
	}

	switch req.Name {
	case "atlas_list_collections":
		return server.ListCollectionsText(ctx)
	case "atlas_search_items":
		collection := stringArg(req.Arguments, "collection", "dataset")
		q := stringArg(req.Arguments, "query", "")
		bbox := stringArg(req.Arguments, "bbox", "")
		limit := intArg(req.Arguments, "limit", 5)
		startindex := intArg(req.Arguments, "startindex", 1)
		return server.SearchItemsText(ctx, atlasclient.SearchParams{
			CollectionID: collection,
			Query:        q,
			BBox:         bbox,
			Limit:        limit,
			StartIndex:   startindex,
		})
	case "atlas_answer_query":
		query := strings.TrimSpace(stringArg(req.Arguments, "query", ""))
		if query == "" {
			return "", fmt.Errorf("atlas_answer_query requires 'query'")
		}
		limit := intArg(req.Arguments, "limit", 5)
		return server.AnswerNaturalQuery(ctx, query, limit)
	default:
		return "", fmt.Errorf("unknown tool: %s", req.Name)
	}
}

func toolDefinitions() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "atlas_list_collections",
			"description": "Lists available Digital Atlas collections.",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			"name":        "atlas_search_items",
			"description": "Searches Digital Atlas items in a collection using query and optional bbox.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"collection": map[string]interface{}{"type": "string", "default": "dataset"},
					"query":      map[string]interface{}{"type": "string"},
					"bbox":       map[string]interface{}{"type": "string", "description": "minLon,minLat,maxLon,maxLat"},
					"limit":      map[string]interface{}{"type": "integer", "default": 5},
					"startindex": map[string]interface{}{"type": "integer", "default": 1},
				},
				"required": []string{"query"},
			},
		},
		{
			"name":        "atlas_answer_query",
			"description": "Answers natural-language location/data questions, e.g. 'Where are the bushfires around Australia?'",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{"type": "string"},
					"limit": map[string]interface{}{"type": "integer", "default": 5},
				},
				"required": []string{"query"},
			},
		},
	}
}

func stringArg(args map[string]interface{}, key, fallback string) string {
	v, ok := args[key]
	if !ok || v == nil {
		return fallback
	}
	s, ok := v.(string)
	if !ok {
		return fallback
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return fallback
	}
	return s
}

func intArg(args map[string]interface{}, key string, fallback int) int {
	v, ok := args[key]
	if !ok || v == nil {
		return fallback
	}
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(t))
		if err == nil {
			return n
		}
	}
	return fallback
}

func readMessage(r *bufio.Reader) ([]byte, error) {
	contentLen := 0
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}

		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "content-length:") {
			raw := strings.TrimSpace(line[len("content-length:"):])
			n, err := strconv.Atoi(raw)
			if err != nil || n < 0 {
				return nil, fmt.Errorf("invalid content-length")
			}
			contentLen = n
		}
	}
	if contentLen <= 0 {
		return nil, fmt.Errorf("missing content-length")
	}

	body := make([]byte, contentLen)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, err
	}
	return body, nil
}

func writeResponse(w io.Writer, resp rpcResponse) error {
	b, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Content-Length: %d\r\n\r\n", len(b)); err != nil {
		return err
	}
	if _, err := w.Write(b); err != nil {
		return err
	}
	return nil
}

func parseID(raw json.RawMessage) (interface{}, error) {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s, nil
	}
	var n float64
	if err := json.Unmarshal(raw, &n); err == nil {
		if n == float64(int64(n)) {
			return int64(n), nil
		}
		return n, nil
	}
	return nil, fmt.Errorf("unsupported id")
}
