package mcpserver

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/whatnick/daa_mcp/internal/atlasclient"
	"github.com/whatnick/daa_mcp/internal/model"
)

const australiaBBox = "109,-45,159,-8"

func (s *Server) ListCollectionsText(ctx context.Context) (string, error) {
	resp, err := s.client.ListCollections(ctx)
	if err != nil {
		return "", err
	}

	if len(resp.Collections) == 0 {
		return "No collections returned by Digital Atlas.", nil
	}

	var b strings.Builder
	b.WriteString("Available collections:\n")
	for i, c := range resp.Collections {
		fmt.Fprintf(&b, "%d. %s (%s)\n", i+1, c.Title, c.ID)
	}
	return strings.TrimSpace(b.String()), nil
}

func (s *Server) SearchItemsText(ctx context.Context, params atlasclient.SearchParams) (string, error) {
	resp, err := s.client.SearchItemsWithParams(ctx, params)
	if err != nil {
		return "", err
	}
	return formatItemsResponse(params, resp), nil
}

func (s *Server) AnswerNaturalQuery(ctx context.Context, question string, limit int) (string, error) {
	params := interpretNaturalQuery(question, limit)
	resp, err := s.client.SearchItemsWithParams(ctx, params)
	if err != nil {
		return "", err
	}

	if len(resp.Features) == 0 {
		return fmt.Sprintf("No matches found for %q.", params.Query), nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Found %d matching items for %q", resp.NumberMatched, params.Query)
	if params.BBox != "" {
		b.WriteString(" within Australia")
	}
	fmt.Fprintf(&b, " (showing %d):\n", len(resp.Features))

	for i, f := range resp.Features {
		title := featureTitle(f)
		kind := firstString(f.Properties, "type")
		if kind == "" {
			kind = "Unknown type"
		}

		fmt.Fprintf(&b, "%d. %s — %s", i+1, title, kind)
		if bbox, ok := featureBBox(f); ok {
			fmt.Fprintf(&b, " — approx bbox %.3f,%.3f to %.3f,%.3f", bbox[0], bbox[1], bbox[2], bbox[3])
		}
		b.WriteString("\n")

		if u := firstString(f.Properties, "url"); u != "" {
			fmt.Fprintf(&b, "   %s\n", u)
		}
	}

	if hasNext, next := nextFromLinks(resp.Links); hasNext {
		fmt.Fprintf(&b, "More results available with startindex=%d.", next)
	}

	return strings.TrimSpace(b.String()), nil
}

func interpretNaturalQuery(question string, limit int) atlasclient.SearchParams {
	raw := strings.TrimSpace(question)
	lower := strings.ToLower(raw)

	q := raw
	replacements := map[string]string{
		"where are the ":   "",
		"where are ":       "",
		"what are the ":    "",
		"what are ":        "",
		"show me ":         "",
		"around australia": "",
		"in australia":     "",
		"across australia": "",
		"etc":              "",
	}
	for old, newVal := range replacements {
		q = strings.ReplaceAll(strings.ToLower(q), old, newVal)
	}
	q = strings.TrimSpace(strings.Trim(q, " ?.,!"))

	switch {
	case strings.Contains(lower, "bushfire"), strings.Contains(lower, "bush fire"), strings.Contains(lower, "wildfire"):
		q = "bushfire"
	case strings.Contains(lower, "flood"):
		q = "flood"
	case strings.Contains(lower, "cyclone"):
		q = "cyclone"
	}
	if q == "" {
		q = raw
	}

	bbox := ""
	if strings.Contains(lower, "australia") || strings.Contains(lower, "around") || strings.Contains(lower, "across") {
		bbox = australiaBBox
	}

	if limit <= 0 {
		limit = 5
	}

	return atlasclient.SearchParams{
		CollectionID: "dataset",
		Query:        q,
		BBox:         bbox,
		Limit:        limit,
		StartIndex:   1,
	}
}

func formatItemsResponse(params atlasclient.SearchParams, resp *model.ItemsResponse) string {
	if len(resp.Features) == 0 {
		return fmt.Sprintf("No items found for %q.", params.Query)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Items for %q in %s: %d matched, %d returned.\n", params.Query, params.CollectionID, resp.NumberMatched, resp.NumberReturned)
	for i, f := range resp.Features {
		title := featureTitle(f)
		fmt.Fprintf(&b, "%d. %s\n", i+1, title)
		if u := firstString(f.Properties, "url"); u != "" {
			fmt.Fprintf(&b, "   %s\n", u)
		}
	}
	if hasNext, next := nextFromLinks(resp.Links); hasNext {
		fmt.Fprintf(&b, "Next page startindex=%d", next)
	}
	return strings.TrimSpace(b.String())
}

func featureTitle(f model.Feature) string {
	if t := firstString(f.Properties, "title"); t != "" {
		return t
	}
	if t := firstString(f.Properties, "name"); t != "" {
		return t
	}
	if t := firstString(f.Properties, "id"); t != "" {
		return t
	}
	return "Untitled"
}

func firstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		v, ok := m[k]
		if !ok || v == nil {
			continue
		}
		switch t := v.(type) {
		case string:
			if strings.TrimSpace(t) != "" {
				return t
			}
		case fmt.Stringer:
			s := strings.TrimSpace(t.String())
			if s != "" {
				return s
			}
		}
	}
	return ""
}

func featureBBox(f model.Feature) ([4]float64, bool) {
	if b, ok := bboxFromValue(f.Geometry); ok {
		return b, true
	}
	if extent, ok := f.Properties["extent"]; ok {
		if b, ok := bboxFromValue(extent); ok {
			return b, true
		}
	}
	return [4]float64{}, false
}

func bboxFromValue(v any) ([4]float64, bool) {
	var coords [][2]float64
	collectCoordinates(v, &coords)
	if len(coords) == 0 {
		return [4]float64{}, false
	}

	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64
	for _, p := range coords {
		if p[0] < minX {
			minX = p[0]
		}
		if p[1] < minY {
			minY = p[1]
		}
		if p[0] > maxX {
			maxX = p[0]
		}
		if p[1] > maxY {
			maxY = p[1]
		}
	}

	return [4]float64{minX, minY, maxX, maxY}, true
}

func collectCoordinates(v any, out *[][2]float64) {
	switch t := v.(type) {
	case map[string]any:
		if coords, ok := t["coordinates"]; ok {
			collectCoordinates(coords, out)
		}
	case []any:
		if len(t) >= 2 {
			x, xOk := toFloat(t[0])
			y, yOk := toFloat(t[1])
			if xOk && yOk {
				*out = append(*out, [2]float64{x, y})
				return
			}
		}
		for _, item := range t {
			collectCoordinates(item, out)
		}
	}
}

func toFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case jsonNumber:
		f, err := strconv.ParseFloat(string(t), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

type jsonNumber string
