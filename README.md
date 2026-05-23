# daa_mcp

`daa_mcp` is a public Go project for a streaming MCP wrapper around the Digital Atlas of Australia Search API (Esri/ArcGIS-backed).

## What this bootstrap includes

- Go module and package structure for:
  - `internal/atlasclient` (HTTP API wrapper)
  - `internal/mcpserver` (streaming event flow)
  - `internal/model` (typed payloads/events)
- A runnable CLI entrypoint at `cmd/daa-mcp` that outputs stream events (`metadata`, `item`, `done`) as JSON lines.
- Copilot guidance in `.github/copilot-instructions.md` aligned to Digital Atlas endpoints and query semantics.

## Upstream API

- Base URL: `https://digital.atlas.gov.au/api/search/v1`
- OpenAPI: `https://digital.atlas.gov.au/api/search/definition/?f=json`
- Known collections:
  - `dataset`
  - `appAndMap`
  - `document`
  - `all`

## Quick start

```bash
go run ./cmd/daa-mcp -collection dataset -q water -limit 5 -startindex 1
```

Set a custom endpoint if needed:

```bash
export DIGITAL_ATLAS_BASE_URL=https://digital.atlas.gov.au/api/search/v1
```

## Roadmap to full MCP server

1. Implement stdio MCP transport and tool registration.
2. Expose endpoint-aligned tools (collections, queryables, items, related, connected, aggregations).
3. Stream paginated results incrementally with cancellation support.
4. Add integration tests with `httptest` and golden fixtures.

## License

MIT

