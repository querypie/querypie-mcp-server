# QueryPie MCP

The MCP server for QueryPie for administrators to manage it.

<a href="https://www.youtube.com/watch?v=uJ8u9oHCiIM">
  <img src="https://raw.githubusercontent.com/querypie/querypie-mcp-server/main/assets/querypie-ai-hub.png" width="60%" alt="QueryPie Agent Demo">
</a>

<br />

## Key Usage Demo

**📊 Dashboard with Instant Charts**

Visualize your data instantly by turning query results into live charts and dashboards—without writing a single line of code.

<img src="https://raw.githubusercontent.com/querypie/querypie-mcp-server/main/assets/chart.gif" width="800" alt="QueryPie Demo">

<br />

**💽 Monitor Disk & Memory Usage**

Track server resource usage like disk space and memory in real time, and identify the most resource-intensive processes at a glance.

<img src="https://raw.githubusercontent.com/querypie/querypie-mcp-server/main/assets/usage.gif" width="800" alt="QueryPie Demo">

<br />

**🕵️‍♂️ Detect Suspicious Behavior**

Review access logs and detect abnormal user behavior, such as risky SQL commands or unauthorized server activity.

<img src="https://raw.githubusercontent.com/querypie/querypie-mcp-server/main/assets/logs.gif" width="800" alt="QueryPie Demo">

<br />

**✨ And That’s Just the Beginning...**

There’s so much more you can do—register assets, manage access, automate audits, and more. 

It all depends on how you use QueryPie.  Start exploring and make it yours.

<br />

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
                "--rm",
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
        "--rm",
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
