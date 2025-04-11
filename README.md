# QueryPie MCP

The MCP server for QueryPie for administrators to manage it.

<a href="https://www.youtube.com/watch?v=nChu-sY9Cu8">
  <img src="https://raw.githubusercontent.com/querypie/querypie-mcp-server/main/assets/querypie-ai-agent-demo.png" width="60%" alt="QueryPie Agent Demo">
</a>

**▶ [Watch English Version](https://www.youtube.com/watch?v=nChu-sY9Cu8)**  
**▶ [日本語版を見る (Watch Japanese Version)](https://www.youtube.com/watch?v=ujtD9UzH-Pw)**

## Installation

Prepare your QueryPie API key and URL.

You can find the API key on <kbd>General</kbd> > <kbd>System</kbd> > <kbd>API Token</kbd> in the QueryPie web console.

### Docker

```bash
# Stdio example
export QUERYPIE_API_KEY=your_token
export QUERYPIE_URL=https://your_querypie_url

docker run --rm \
    -e "QUERYPIE_API_KEY=${QUERYPIE_API_KEY}" \
    ghcr.io/querypie/querypie-mcp-server "${QUERYPIE_URL}"
```

```bash
# SSE example
export QUERYPIE_API_KEY=your_token
export QUERYPIE_URL=https://your_querypie_url

docker run --rm \
    -e "QUERYPIE_API_KEY=${QUERYPIE_API_KEY}" \
    ghcr.io/querypie/querypie-mcp-server "${QUERYPIE_URL}" \
    --transport sse \
    --port 8000
```

### Linux/macOS

```bash
# Install the querypie-mcp-server binary to ~/.local/bin
curl -L https://github.com/querypie/querypie-mcp-server/releases/latest/download/install.sh | sh
```

```bash
# Stdio example
export QUERYPIE_API_KEY=your_token
querypie-mcp-server https://your_querypie_url
```

```bash
# SSE example
export QUERYPIE_API_KEY=your_token
querypie-mcp-server https://your_querypie_url \
    --transport sse \
    --port 8000
```

### Claude Desktop

Add this into your `claude_desktop_config.json` (either at `~/Library/Application Support/Claude` on macOS or `C:\Users\NAME\AppData\Roaming\Claude` on Windows):

```json
{
    "mcpServers": {
        "querypie": {
            "command": "docker",
            "args": [
                "run",
                "-e",
                "QUERYPIE_API_KEY=${QUERYPIE_API_KEY}",
                "-it",
                "ghcr.io/querypie/querypie-mcp-server",
                "https://your_querypie_url"
            ],
            "env": {
                "QUERYPIE_API_KEY": "your_token"
            }
        }
    }
}
```

### Cursor

Add this into your `~/cursor/mcp.json`

```json
{
  "mcpServers": {
    "querypie": {
      "command": "docker",
      "type": "stdio",
      "args": [
        "run",
        "-e",
        "QUERYPIE_API_KEY=${QUERYPIE_API_KEY}",
        "-it",
        "ghcr.io/querypie/querypie-mcp-server",
        "https://your_querypie_url"
      ],
      "env": {
        "QUERYPIE_API_KEY": "your_token"
      }
    }
  }
}
```
