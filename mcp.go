package simplemcp

import (
	"context"
	"encoding/json"

	"github.com/sourcegraph/jsonrpc2"
)

type ToolFunc func(ctx context.Context, params json.RawMessage) (CallToolResult, error)

type ToolImpl struct {
	Func ToolFunc
}

func NewHandler(serverInfo *Implementation) *Handler {
	return &Handler{
		serverInfo: serverInfo,
		tools:      []*Tool{},
		toolImpls:  map[string]*ToolImpl{},
	}
}

type Handler struct {
	serverInfo *Implementation
	tools      []*Tool
	toolImpls  map[string]*ToolImpl
}

type RegisterToolConfig struct {
	Name        string
	Description string
	Properties  map[string]Property
	Required    []string
	ToolFunc    ToolFunc
}

func (h *Handler) RegisterTool(c *RegisterToolConfig) {
	h.tools = append(h.tools, &Tool{
		Name:        c.Name,
		Description: c.Description,
		InputSchema: InputSchema{
			Type:       InputSchemaTypeObject,
			Properties: c.Properties,
			Required:   c.Required,
		},
	})
	h.toolImpls[c.Name] = &ToolImpl{
		Func: c.ToolFunc,
	}
}

func (h *Handler) Handle(
	ctx context.Context,
	conn *jsonrpc2.Conn,
	req *jsonrpc2.Request,
) {
	switch req.Method {
	case InitializeMethod:
		var params InitializeParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
				Code:    jsonrpc2.CodeInvalidParams,
				Message: err.Error(),
			})
			return
		}
		result := InitializeResult{
			ProtocolVersion: params.ProtocolVersion,
			Capabilities: ServerCapabilities{
				Tools: &Tools{},
			},
			ServerInfo: *h.serverInfo,
		}
		conn.Reply(ctx, req.ID, result)
	case NotificationsInitializedMethod:
		return
	case ToolsListMethod:
		result := ListToolsResult{
			Tools: h.tools,
		}
		conn.Reply(ctx, req.ID, result)
	case ToolsCallMethod:
		var params CallToolParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
				Code:    jsonrpc2.CodeInvalidParams,
				Message: err.Error(),
			})
			return
		}
		impl, ok := h.toolImpls[params.Name]
		if !ok {
			conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
				Code:    jsonrpc2.CodeInvalidParams,
				Message: "tool not found",
			})
			return
		}
		result, err := impl.Func(ctx, params.Arguments)
		if err != nil {
			conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
				Code:    jsonrpc2.CodeInternalError,
				Message: err.Error(),
			})
			return
		}
		conn.Reply(ctx, req.ID, result)
	default:
		conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeMethodNotFound,
			Message: "method not found",
		})
	}
}

func (h *Handler) Run(ctx context.Context) {
	stdio := newStdioReadWriteCloser()
	stream := jsonrpc2.NewPlainObjectStream(stdio)

	conn := jsonrpc2.NewConn(ctx, stream, h)
	defer conn.Close()
	<-conn.DisconnectNotify()
}
