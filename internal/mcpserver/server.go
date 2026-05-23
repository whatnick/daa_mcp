package mcpserver

import (
	"context"

	"github.com/whatnick/daa_mcp/internal/atlasclient"
	"github.com/whatnick/daa_mcp/internal/model"
)

type Server struct {
	client *atlasclient.Client
}

func New(client *atlasclient.Client) *Server {
	return &Server{client: client}
}

// StreamSearchItems emits predictable stream events for MCP tool wrappers:
// metadata -> item* -> done.
func (s *Server) StreamSearchItems(
	ctx context.Context,
	collectionID string,
	q string,
	limit int,
	startIndex int,
	emit func(model.StreamEvent) error,
) error {
	resp, err := s.client.SearchItems(ctx, collectionID, q, limit, startIndex)
	if err != nil {
		return err
	}

	if err := emit(model.StreamEvent{
		Event: "metadata",
		Metadata: &model.SearchMetadata{
			CollectionID:   collectionID,
			Query:          q,
			NumberMatched:  resp.NumberMatched,
			NumberReturned: resp.NumberReturned,
			Limit:          limit,
			StartIndex:     startIndex,
		},
	}); err != nil {
		return err
	}

	emitted := 0
	for i := range resp.Features {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		item := resp.Features[i]
		if err := emit(model.StreamEvent{
			Event: "item",
			Item:  &item,
		}); err != nil {
			return err
		}
		emitted++
	}

	hasNext, nextStart := nextFromLinks(resp.Links)
	return emit(model.StreamEvent{
		Event: "done",
		Done: &model.SearchDone{
			EmittedCount:   emitted,
			HasNextPage:    hasNext,
			NextStartIndex: nextStart,
		},
	})
}

func nextFromLinks(links []model.Link) (bool, int) {
	for _, link := range links {
		if link.Rel != "next" {
			continue
		}
		return true, extractStartIndex(link.Href)
	}
	return false, 0
}

func extractStartIndex(rawURL string) int {
	const key = "startindex="
	idx := -1
	for i := 0; i+len(key) <= len(rawURL); i++ {
		if rawURL[i:i+len(key)] == key {
			idx = i + len(key)
			break
		}
	}
	if idx < 0 {
		return 0
	}
	n := 0
	for idx < len(rawURL) {
		ch := rawURL[idx]
		if ch < '0' || ch > '9' {
			break
		}
		n = n*10 + int(ch-'0')
		idx++
	}
	return n
}
