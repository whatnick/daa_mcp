# daa_mcp

`daa_mcp` is a public Go MCP server wrapper around the Digital Atlas of Australia Search API (Esri/ArcGIS-backed).

## What this includes

- Go module and package structure for:
  - `internal/atlasclient` (HTTP API wrapper)
  - `internal/mcpserver` (streaming event flow)
  - `internal/model` (typed payloads/events)
- A runnable **stdio MCP server** entrypoint at `cmd/daa-mcp`.
- MCP tools:
  - `atlas_list_collections`
  - `atlas_search_items`
  - `atlas_answer_query` (natural-language query answering, e.g. bushfires around Australia)
- Multi-agent setup docs in `docs/AGENTS.md` (Claude, VS Code Copilot, Kiro).
- Demo video pipeline docs in `docs/DEMO_VIDEO.md`.
- Local voice cloning guidance in `docs/VOICE_CLONING_LOCAL.md`.
- Copilot guidance in `.github/copilot-instructions.md` aligned to Digital Atlas endpoints and query semantics.

## Upstream API

- Base URL: `https://digital.atlas.gov.au/api/search/v1`
- OpenAPI: `https://digital.atlas.gov.au/api/search/definition/?f=json`
- Known collections:
  - `dataset`
  - `appAndMap`
  - `document`
  - `all`

## Run as MCP server

```bash
go run ./cmd/daa-mcp
```

Set a custom endpoint if needed:

```bash
export DIGITAL_ATLAS_BASE_URL=https://digital.atlas.gov.au/api/search/v1
```

## Example MCP tool call

Use `tools/call` with:

- `name`: `atlas_answer_query`
- `arguments.query`: `"Where are the bushfires around Australia?"`
- `arguments.limit`: `5`

The tool returns a concise text answer with matching datasets/maps, approximate spatial bounds, and source URLs.

## Build and publish binaries

Tagging a release (for example `v0.2.0`) triggers `.github/workflows/release.yml`, which:

1. runs `go test ./...`
2. builds cross-platform binaries (`linux/darwin/windows` for `amd64` and `arm64`)
3. packages archives and publishes SHA256 `checksums.txt`
4. publishes `mcp-manifest.json` and MCP client example config files as release assets

Create and push a tag:

```bash
git tag v0.2.0
git push origin v0.2.0
```

## MCP metadata and client templates

- Manifest template: `mcp/manifest.template.json`
- Client examples:
  - `mcp/clients/claude_desktop_config.example.json`
  - `mcp/clients/vscode.mcp.example.json`
  - `mcp/clients/kiro.mcp.example.json`

## Demo video (headless)

```bash
npm install
npx playwright install chromium
npm run demo:record
```

See:

- `docs/DEMO_VIDEO.md`
- Latest release demo asset: [hazards-demo.webm](https://github.com/whatnick/daa_mcp/releases/latest/download/hazards-demo.webm)

## License

MIT
