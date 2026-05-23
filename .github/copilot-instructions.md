# Copilot instructions for Digital Atlas (Esri) MCP wrapper

This repository builds a **streaming MCP server in Go** that wraps the Digital Atlas of Australia search API (ArcGIS/Esri-backed).

## Scope and architecture

- Implement as a Go MCP server using streaming responses for long result sets.
- Keep business logic separated:
  1. `internal/atlasclient` for HTTP/API calls
  2. `internal/mcpserver` for MCP tool registration and streaming handlers
  3. `internal/model` for response/query types
- All outbound requests must be context-aware (`http.NewRequestWithContext`) and cancellable.

## Source API to wrap

Use `https://digital.atlas.gov.au/api/search/v1` as default base URL.

Primary endpoints:

- `GET /collections`
- `GET /collections/{collectionId}`
- `GET /collections/{collectionId}/queryables`
- `GET /collections/{collectionId}/items`
- `GET /collections/{collectionId}/items/{itemId}`
- `GET /collections/{collectionId}/items/{recordId}/related`
- `GET /collections/{collectionId}/items/{recordId}/connected`
- `GET /collections/{collectionId}/aggregations`
- `GET /conformance`
- OpenAPI definition: `https://digital.atlas.gov.au/api/search/definition/?f=json`

Known collection IDs:

- `dataset`
- `appAndMap`
- `document`
- `all`

## Query parameter conventions

Preserve exact upstream parameter names (do not camelCase them in the wire format):

- `q`
- `bbox` (4 comma-separated coordinates)
- `filter` (OGC CQL style filter expression)
- `limit`
- `startindex`
- `type`
- `title`
- `recordId`
- `sortBy`
- `tags`
- `openData`
- `token` (optional ArcGIS token)

Always URL-encode filter expressions and string values safely.

## MCP tool design

Expose thin, predictable MCP tools aligned to API resources:

1. `atlas_list_collections`
2. `atlas_get_collection`
3. `atlas_get_queryables`
4. `atlas_search_items`
5. `atlas_get_item`
6. `atlas_get_related_items`
7. `atlas_get_connected_items`
8. `atlas_get_aggregations`

For item-search style tools, support pagination and streaming:

- Accept `limit` + `startindex`.
- Return summary first (`numberMatched`, `numberReturned`).
- Stream item batches (or per-item events) before final completion payload.

## Streaming behavior requirements

- Use deterministic event sequencing:
  1. `metadata` event
  2. one or more `item_batch` (or `item`) events
  3. `done` event
- Include continuation hints (`next_startindex`) when `next` links are present.
- Do not buffer all items before emitting output.
- Respect context cancellation immediately and stop upstream calls.

## Reliability and error handling

- Use a shared `http.Client` with sane timeouts.
- Handle non-2xx responses as typed errors with status code and endpoint context.
- Surface upstream JSON decoding errors explicitly; do not swallow partial failures.
- Add retry only for transient transport/server failures (5xx, network reset, timeout), with bounded exponential backoff.

## Data mapping rules

- Preserve raw upstream JSON fields in typed structs where practical.
- For features returned by `/items`, map and expose:
  - top-level: `type`, `numberMatched`, `numberReturned`, `links`
  - feature-level: `id`, `geometry`, `properties`
- Keep unknown JSON fields with `map[string]any` when schema varies across collections.

## Testing expectations

- Use table-driven tests for query construction and endpoint routing.
- Add httptest-based tests for:
  - pagination behavior (`startindex`, `limit`)
  - streaming event ordering
  - cancellation propagation
  - error mapping for 4xx/5xx and invalid JSON
- Include golden JSON fixtures for representative `dataset` and `appAndMap` results.

## Implementation constraints

- Keep dependencies minimal; prefer Go standard library first.
- No global mutable state for request parameters or stream cursors.
- Log request IDs and endpoint paths, but never log tokens.
- Make base URL configurable via env var, defaulting to:
  - `DIGITAL_ATLAS_BASE_URL=https://digital.atlas.gov.au/api/search/v1`
