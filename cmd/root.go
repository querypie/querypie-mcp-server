package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/cobra"

	"github.com/querypie/querypie-mcp-server/internal/consts"
	"github.com/querypie/querypie-mcp-server/internal/tools"
	"github.com/querypie/querypie-mcp-server/server"
)

var (
	transportFlag string
	portFlag      int
	noCacheFlag   bool
)

var rootCmd = &cobra.Command{
	Use:     "mcp-querypie <querypie-url>",
	Short:   "Run the MCP Server for QueryPie",
	Long:    `Run the MCP Server for QueryPie.`,
	Version: consts.Version,
	Example: `  QUERYPIE_API_KEY=ap111111 mcp-querypie https://api.querypie.com --transport stdio
  QUERYPIE_API_KEY=ap111111 mcp-querypie https://api.querypie.com --transport sse --port 8000`,
	Args: cobra.MatchAll(func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("argument <querypie-url> is required")
		}

		if len(args) > 1 {
			return fmt.Errorf("only one argument <querypie-url> is allowed")
		}

		return nil
	}),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
		defer stop()

		// Check environment varaibles
		querypieAPIKey := os.Getenv("QUERYPIE_API_KEY")
		if querypieAPIKey == "" {
			return errors.New("QUERYPIE_API_KEY is not set")
		} else if len(querypieAPIKey) != 38 && strings.HasPrefix(querypieAPIKey, "ap") {
			return errors.New("malformed QUERYPIE_API_KEY. please check the API key")
		}

		// Check positional arguments
		if len(args) == 0 {
			return fmt.Errorf("querypie-url is required")
		}

		// Check flags
		transport := transportFlag
		if transport != "stdio" && transport != "sse" {
			return fmt.Errorf("invalid transport: %s", transport)
		}

		port := portFlag
		if port < 0 || port > 65535 {
			return fmt.Errorf("invalid port: %d", port)
		}

		server := server.NewServer(querypieAPIKey, args[0], transport, port, server.NewPromptServerOptions()...)
		return server.Start(ctx, noCacheFlag)
	},
}

func init() {
	// Set up logging first
	logLevel := slog.LevelInfo
	if tools.IsDev() {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	// Set up command flags
	rootCmd.Flags().StringVarP(&transportFlag, "transport", "t", "stdio", "Transport mode (stdio|sse)")
	rootCmd.Flags().IntVarP(&portFlag, "port", "p", 8000, "Port number if transport is sse")
	rootCmd.Flags().BoolVarP(&noCacheFlag, "no-cache", "n", false, "Do not cache the OpenAPI specification")
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
