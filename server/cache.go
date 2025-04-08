package server

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/querypie/querypie-mcp-server/internal/consts"
)

const CacheTtl = time.Hour * 12

var ErrCacheFileOutdated = fmt.Errorf("cache file is outdated")

func loadCachedOpenAPI(version Version) ([]byte, error) {
	cacheDir := filepath.Join(os.TempDir(), ".mcp-querypie", version.String())
	os.MkdirAll(cacheDir, 0755)
	cacheFile := filepath.Join(cacheDir, "openapi.yaml")

	if fileInfo, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("openapi.yaml file does not exist in cache: %w", err)
	} else if fileInfo.ModTime().Before(time.Now().Add(-CacheTtl)) {
		return nil, ErrCacheFileOutdated
	}

	return os.ReadFile(cacheFile)
}

func writeOpenAPIToCache(version Version, spec []byte) error {
	cacheDir := filepath.Join(os.TempDir(), ".mcp-querypie", version.String())
	os.MkdirAll(cacheDir, 0755)
	cacheFile := filepath.Join(cacheDir, "openapi.yaml")
	return os.WriteFile(cacheFile, spec, 0644)
}

func downloadOpenAPIFile(version Version, depth int) ([]byte, error) {
	repositoryURL := "https://github.com/querypie/querypie-mcp-server"
	openapiURL := fmt.Sprintf("%s/releases/download/v%s/%s-openapi.yaml", repositoryURL, consts.Version, strings.ReplaceAll(version.String(), ".", "-"))
	response, err := http.Get(openapiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get openapi.yaml from %s: %w", openapiURL, err)
	}
	defer response.Body.Close()

	spec, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read openapi.yaml from %s: %w", openapiURL, err)
	}

	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusNotFound && depth < 1 {
			slog.Warn("• OpenAPI specification not found in the release. Fallback to v10.2.0")
			return downloadOpenAPIFile(Version{"10", "2", "0"}, depth+1)
		}

		return nil, fmt.Errorf("failed to get openapi.yaml from %s: %s", openapiURL, response.Status)
	}

	return spec, nil
}
