package mcpserver

import "testing"

func TestInterpretNaturalQueryBushfiresAustralia(t *testing.T) {
	params := interpretNaturalQuery("Where are the bushfires around Australia?", 3)

	if params.Query != "bushfire" {
		t.Fatalf("expected query bushfire, got %q", params.Query)
	}
	if params.BBox != australiaBBox {
		t.Fatalf("expected australia bbox, got %q", params.BBox)
	}
	if params.CollectionID != "dataset" {
		t.Fatalf("expected dataset collection, got %q", params.CollectionID)
	}
	if params.Limit != 3 {
		t.Fatalf("expected limit 3, got %d", params.Limit)
	}
}

func TestBboxFromValue(t *testing.T) {
	v := map[string]any{
		"type": "Polygon",
		"coordinates": []any{
			[]any{
				[]any{109.7, -45.5},
				[]any{159.4, -45.5},
				[]any{159.4, -6.8},
				[]any{109.7, -6.8},
			},
		},
	}

	b, ok := bboxFromValue(v)
	if !ok {
		t.Fatalf("expected bbox")
	}
	if b[0] != 109.7 || b[1] != -45.5 || b[2] != 159.4 || b[3] != -6.8 {
		t.Fatalf("unexpected bbox: %#v", b)
	}
}
