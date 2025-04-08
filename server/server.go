package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/mark3labs/mcp-go/server"
	"github.com/pb33f/libopenapi"

	"github.com/querypie/querypie-mcp-server/internal/consts"
)

type Server struct {
	querypieAPIKey string
	querypieURL    string
	transport      string
	port           int
	opts           []server.ServerOption
}

func NewServer(querypieAPIKey string, querypieURL string, transport string, port int, opts ...server.ServerOption) *Server {
	return &Server{
		querypieAPIKey: querypieAPIKey,
		querypieURL:    querypieURL,
		transport:      transport,
		port:           port,
		opts:           opts,
	}
}

func (s *Server) Start(ctx context.Context, noCache bool) error {
	slog.Info("• Starting MCP Server for QueryPie")

	// getting version from the querypie server
	slog.Info("• Getting version from the QueryPie", "url", s.querypieURL)
	version, err := getVersion(s.querypieURL)
	if err != nil {
		return fmt.Errorf("failed to get version from %s: %w", s.querypieURL, err)
	}
	slog.Info(fmt.Sprintf("   ✔ QueryPie version is resolved: %s", version.String()))

	slog.Info("• Loading OpenAPI specification")

	var spec []byte

	if noCache {
		slog.Info("   • OpenAPI specification is not cached. Downloading new one")
		spec, err = downloadOpenAPIFile(*version, 0)
		if err != nil {
			return err
		}
		slog.Info("   ✔ OpenAPI specification is downloaded")
	} else {
		cachedOpenAPIFile, err := loadCachedOpenAPI(*version)
		if err != nil {
			if errors.Is(err, ErrCacheFileOutdated) {
				slog.Info("   • OpenAPI specification is outdated. Downloading new one")
			} else if os.IsNotExist(err) {
				slog.Info("   • OpenAPI specification is not cached. Downloading new one")
			} else {
				slog.Debug("   • Failed to load cached openapi.yaml. Downloading new one", "error", err)
			}

			spec, err = downloadOpenAPIFile(*version, 0)
			if err != nil {
				return err
			}

			err = writeOpenAPIToCache(*version, spec)
			if err != nil {
				slog.Debug("failed to write openapi.yaml cache", "error", err)
			}

			slog.Info("   ✔ OpenAPI specification is downloaded")
		} else {
			spec = cachedOpenAPIFile
			slog.Info("   ✔ OpenAPI specification is loaded from cache")
		}
	}

	doc, err := libopenapi.NewDocument(spec)
	if err != nil {
		return fmt.Errorf("error parsing OpenAPI spec: %v", err)
	}

	model, errs := doc.BuildV3Model()
	if len(errs) > 0 {
		return fmt.Errorf("error building OpenAPI model: %v", errors.Join(errs...))
	}

	slog.Info("   • Loading tools from OpenAPI specification")
	tools, err := parseToolsFromOpenAPI(ctx, s.querypieAPIKey, s.querypieURL, model.Model)
	if err != nil {
		return fmt.Errorf("error parsing tools from OpenAPI: %w", err)
	}
	slog.Info(fmt.Sprintf("   ✔ %d tools are loaded", len(tools)))

	var opts []server.ServerOption
	opts = append(opts, server.WithLogging())
	opts = append(opts, s.opts...)
	srv := server.NewMCPServer("mcp-querypie", consts.Version, opts...)

	srv.AddTools(tools...)

	switch s.transport {
	case "stdio":
		slog.Info(fmt.Sprintf("✔ MCP Server is started with %s", s.transport))
		stdioSrv := server.NewStdioServer(srv)
		return stdioSrv.Listen(ctx, os.Stdin, os.Stdout)
	case "sse":
		slog.Info(fmt.Sprintf("✔ MCP Server with %s is now listening on :%d", s.transport, s.port))
		sseSrv := server.NewSSEServer(srv)
		errChan := make(chan error)
		go func() {
			err := sseSrv.Start(fmt.Sprintf(":%d", s.port))
			if err != nil {
				errChan <- err
			}
		}()

		select {
		case <-ctx.Done():
			slog.Info("• Shutting down MCP Server ...")
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			err := sseSrv.Shutdown(ctx)
			slog.Info("• MCP Server is shutdown")
			if errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return err
		case err := <-errChan:
			return err
		}
	default:
		return fmt.Errorf("unsupported transport: %s", s.transport)
	}
}

type Version struct {
	Major string
	Minor string
	Patch string
}

func (v Version) String() string {
	return fmt.Sprintf("v%s.%s.%s", v.Major, v.Minor, v.Patch)
}

func getVersion(querypieURL string) (*Version, error) {
	versionURL, err := url.JoinPath(querypieURL, "/version")
	if err != nil {
		return nil, fmt.Errorf("malformed querypie URL: %w", err)
	}

	response, err := http.Get(versionURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get version from %s: %w", querypieURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get version from %s: %s", querypieURL, response.Status)
	}

	var body struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to read version from %s: %w", querypieURL, err)
	}

	version := body.Version
	re := regexp.MustCompile(`^([0-9]+).([0-9]+).([0-9]+)`)
	matches := re.FindStringSubmatch(version)
	if len(matches) != 4 {
		return nil, fmt.Errorf("failed to parse version from %s. (version=%s)", querypieURL, version)
	}

	return &Version{
		Major: matches[1],
		Minor: matches[2],
		Patch: matches[3],
	}, nil
}
