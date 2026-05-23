# Using `daa_mcp` with Claude, VS Code Copilot, and Kiro

This server is a **single monolithic stdio binary** (no sidecar services required).

## 1. Launch locally

From source:

```bash
go run ./cmd/daa-mcp
```

From a release binary:

```bash
./daa-mcp
```

Optional environment variable:

```bash
export DIGITAL_ATLAS_BASE_URL=https://digital.atlas.gov.au/api/search/v1
```

## 2. Claude Desktop

Add/update Claude Desktop MCP config:

- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Windows: `%APPDATA%\\Claude\\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "daa_mcp": {
      "command": "/absolute/path/to/daa-mcp",
      "args": [],
      "env": {
        "DIGITAL_ATLAS_BASE_URL": "https://digital.atlas.gov.au/api/search/v1"
      }
    }
  }
}
```

## 3. VS Code Copilot (MCP server config)

Use workspace MCP config at `.vscode/mcp.json`:

```json
{
  "servers": {
    "daa_mcp": {
      "type": "stdio",
      "command": "/absolute/path/to/daa-mcp",
      "args": [],
      "env": {
        "DIGITAL_ATLAS_BASE_URL": "https://digital.atlas.gov.au/api/search/v1"
      }
    }
  }
}
```

## 4. Kiro

Kiro MCP setup follows the same stdio launch pattern: set a server entry with:

- command: absolute path to `daa-mcp`
- args: `[]`
- env: `DIGITAL_ATLAS_BASE_URL`

Example config block (for Kiro MCP settings):

```json
{
  "mcpServers": {
    "daa_mcp": {
      "command": "/absolute/path/to/daa-mcp",
      "args": [],
      "env": {
        "DIGITAL_ATLAS_BASE_URL": "https://digital.atlas.gov.au/api/search/v1"
      }
    }
  }
}
```

## 5. Tools exposed by this MCP

- `atlas_list_collections`
- `atlas_search_items`
- `atlas_answer_query`

Example query:

`Where are the bushfires around Australia?`

## 6. Release assets and metadata

Each tagged release publishes:

- platform binaries (`linux`, `darwin`, `windows`)
- SHA256 checksums
- `mcp-manifest.json`
- client config examples

Use these assets to install and wire the server into MCP clients without building from source.

