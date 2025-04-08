package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

type SchemaType string

const (
	SchemaTypeString  SchemaType = "string"
	SchemaTypeInteger SchemaType = "integer"
	SchemaTypeBoolean SchemaType = "boolean"
	SchemaTypeArray   SchemaType = "array"
	SchemaTypeObject  SchemaType = "object"
)

func parseToolsFromOpenAPI(ctx context.Context, querypieAPIKey, querypieURL string, model v3.Document) ([]server.ServerTool, error) {
	tools := []server.ServerTool{}

	serverURL, err := url.Parse(querypieURL)
	if err != nil {
		return nil, fmt.Errorf("malformed querypie URL: %w", err)
	}

	for pair := model.Paths.PathItems.First(); pair != nil; pair = pair.Next() {
		pathKey := pair.Key()
		pathItem := pair.Value()

		operations := []struct {
			method string
			op     *v3.Operation
		}{
			{"GET", pathItem.Get},
			{"POST", pathItem.Post},
			{"PUT", pathItem.Put},
			{"DELETE", pathItem.Delete},
			{"PATCH", pathItem.Patch},
		}

		for _, op := range operations {
			if op.op == nil || op.op.OperationId == "" {
				continue
			}
			operationID := op.op.OperationId

			var toolOpts []mcp.ToolOption

			// Add operation's descriptions
			if op.op.Description != "" {
				toolOpts = append(toolOpts, mcp.WithDescription(op.op.Description))
			} else {
				toolOpts = append(toolOpts, mcp.WithDescription(op.op.Summary))
			}

			// Add path parameters
			for _, param := range pathItem.Parameters {
				if param == nil || param.Schema == nil {
					continue
				}
				toolOpts = append(toolOpts, convertParamToToolOption(param))
			}

			// Add operation parameters
			for _, param := range op.op.Parameters {
				if param == nil || param.Schema == nil {
					continue
				}
				toolOpts = append(toolOpts, convertParamToToolOption(param))
			}

			// Add request body if present
			if op.op.RequestBody != nil && op.op.RequestBody.Content != nil {
				if mediaType, ok := op.op.RequestBody.Content.Get("application/json"); ok && mediaType != nil {
					if mediaType.Schema != nil && mediaType.Schema.Schema() != nil {
						schema := mediaType.Schema.Schema()

						// flattening the schema
						if schema.Properties != nil {
							for propPair := schema.Properties.First(); propPair != nil; propPair = propPair.Next() {
								schemaType, promptOpts := convertSchemaToToolOption(propPair.Value().Schema())
								switch schemaType {
								case SchemaTypeString:
									toolOpts = append(toolOpts, mcp.WithString(propPair.Key(), promptOpts...))
								case SchemaTypeInteger:
									toolOpts = append(toolOpts, mcp.WithNumber(propPair.Key(), promptOpts...))
								case SchemaTypeBoolean:
									toolOpts = append(toolOpts, mcp.WithBoolean(propPair.Key(), promptOpts...))
								case SchemaTypeArray:
									toolOpts = append(toolOpts, mcp.WithArray(propPair.Key(), promptOpts...))
								case SchemaTypeObject:
									toolOpts = append(toolOpts, mcp.WithObject(propPair.Key(), promptOpts...))
								}
							}
						}

					}
				}
			}

			tools = append(tools, server.ServerTool{
				Tool: mcp.NewTool(operationID, toolOpts...),
				Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
					u := *serverURL
					u.Path = path.Join(u.Path, pathKey)

					headers := make(http.Header)

					// Handle path params
					for _, param := range pathItem.Parameters {
						if param == nil {
							continue
						}

						if value, ok := request.Params.Arguments[param.Name]; ok {
							switch param.In {
							case "path":
								u.Path = strings.ReplaceAll(u.Path, fmt.Sprintf("{%s}", param.Name), url.PathEscape(fmt.Sprint(value)))
							case "query":
								switch v := value.(type) {
								case []interface{}:
									values := make([]string, len(v))
									for i, item := range v {
										values[i] = fmt.Sprint(item)
									}
									u.Query().Add(param.Name, strings.Join(values, ","))
								default:
									u.Query().Add(param.Name, fmt.Sprint(value))
								}
							case "header":
								headers.Add(param.Name, fmt.Sprint(value))
							}
						}
					}

					// Handle operation params
					for _, param := range op.op.Parameters {
						if param == nil {
							continue
						}

						if value, ok := request.Params.Arguments[param.Name]; ok {
							switch param.In {
							case "path":
								u.Path = strings.ReplaceAll(u.Path, fmt.Sprintf("{%s}", param.Name), url.PathEscape(fmt.Sprint(value)))
							case "query":
								u.Query().Add(param.Name, fmt.Sprint(value))
							case "header":
								headers.Add(param.Name, fmt.Sprint(value))
							}
						}
					}

					// Handle request body
					body := make(map[string]interface{})
					if op.op.RequestBody != nil && op.op.RequestBody.Content != nil {
						if mediaType, ok := op.op.RequestBody.Content.Get("application/json"); ok && mediaType != nil {
							if mediaType.Schema != nil && mediaType.Schema.Schema() != nil {
								schema := mediaType.Schema.Schema()
								if schema.Properties != nil {
									for pair := schema.Properties.First(); pair != nil; pair = pair.Next() {
										if value, ok := request.Params.Arguments[pair.Key()]; ok {
											body[pair.Key()] = value
										}
									}
								}
							}
						}
					}

					var reqBody io.Reader
					if len(body) > 0 {
						jsonBody, err := json.Marshal(body)
						if err != nil {
							return nil, fmt.Errorf("failed to marshal request body: %w", err)
						}
						reqBody = bytes.NewReader(jsonBody)
					}

					req, err := http.NewRequest(op.method, u.String(), reqBody)
					if err != nil {
						return nil, fmt.Errorf("failed to create request: %w", err)
					}

					req.Header = headers.Clone()
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Authorization", "Bearer "+querypieAPIKey)

					resp, err := http.DefaultClient.Do(req)
					if err != nil {
						return nil, fmt.Errorf("failed to send request: %w", err)
					}

					defer resp.Body.Close()

					bodyBytes, err := io.ReadAll(resp.Body)
					if err != nil {
						return nil, fmt.Errorf("failed to read response body: %w", err)
					}

					var result *mcp.CallToolResult
					if resp.StatusCode >= 400 {
						result = &mcp.CallToolResult{
							Result: mcp.Result{},
							Content: []mcp.Content{
								mcp.NewTextContent(string(bodyBytes)),
							},
							IsError: true,
						}
					} else {
						result = &mcp.CallToolResult{
							Result: mcp.Result{},
							Content: []mcp.Content{
								mcp.NewTextContent(string(bodyBytes)),
							},
							IsError: false,
						}
					}
					return result, nil
				},
			})
		}
	}

	return tools, nil
}

func convertParamToToolOption(param *v3.Parameter) mcp.ToolOption {
	paramName := param.Name
	schema := param.Schema.Schema()

	schemaType, promptOpts := convertSchemaToToolOption(schema)

	desc := param.Description
	if schemaDesc := buildSchemaDescription(schema); schemaDesc != desc {
		desc += "\n\n" + schemaDesc
	}
	promptOpts = append(promptOpts, mcp.Description(strings.TrimSpace(desc)))

	if param.Required != nil && *param.Required {
		promptOpts = append(promptOpts, mcp.Required())
	}
	switch schemaType {
	case SchemaTypeString:
		return mcp.WithString(paramName, promptOpts...)
	case SchemaTypeInteger:
		return mcp.WithNumber(paramName, promptOpts...)
	case SchemaTypeBoolean:
		return mcp.WithBoolean(paramName, promptOpts...)
	case SchemaTypeArray:
		return mcp.WithArray(paramName, promptOpts...)
	case SchemaTypeObject:
		return mcp.WithObject(paramName, promptOpts...)
	}
	panic("unsupported schema type: " + schemaType)
}

func convertSchemaToToolOption(schema *base.Schema) (SchemaType, []mcp.PropertyOption) {
	var promptOpts []mcp.PropertyOption
	schemaType := SchemaTypeString
	if schema != nil {
		if len(schema.Type) > 0 {
			schemaType = SchemaType(schema.Type[0])
		}
		if schema.Pattern != "" {
			promptOpts = append(promptOpts, mcp.Pattern(schema.Pattern))
		}
		promptOpts = append(promptOpts, mcp.Description(buildSchemaDescription(schema)))
	}
	return schemaType, promptOpts
}

func buildSchemaDescription(schema *base.Schema) string {
	sb := strings.Builder{}

	if schema.Description != "" {
		sb.WriteString(schema.Description)
		sb.WriteString("\n\n")
	}

	if len(schema.Enum) > 0 {
		sb.WriteString("Enum values:\n")
		for _, enum := range schema.Enum {
			sb.WriteString(fmt.Sprintf("- %s\n", enum.Value))
		}
		sb.WriteString("\n\n")
	}

	return strings.TrimSpace(sb.String())
}
