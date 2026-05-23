package model

import "encoding/json"

type CollectionsResponse struct {
	Collections []Collection `json:"collections"`
}

type Collection struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type ItemsResponse struct {
	Type           string    `json:"type"`
	NumberMatched  int       `json:"numberMatched"`
	NumberReturned int       `json:"numberReturned"`
	Links          []Link    `json:"links"`
	Features       []Feature `json:"features"`
	Timestamp      string    `json:"timestamp,omitempty"`
	Raw            json.RawMessage
}

type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

type Feature struct {
	Type       string         `json:"type"`
	Geometry   map[string]any `json:"geometry"`
	Properties map[string]any `json:"properties"`
}

type StreamEvent struct {
	Event string `json:"event"`

	Metadata *SearchMetadata `json:"metadata,omitempty"`
	Item     *Feature        `json:"item,omitempty"`
	Done     *SearchDone     `json:"done,omitempty"`
}

type SearchMetadata struct {
	CollectionID   string `json:"collectionId"`
	Query          string `json:"query,omitempty"`
	NumberMatched  int    `json:"numberMatched"`
	NumberReturned int    `json:"numberReturned"`
	Limit          int    `json:"limit"`
	StartIndex     int    `json:"startindex"`
}

type SearchDone struct {
	EmittedCount   int  `json:"emittedCount"`
	HasNextPage    bool `json:"hasNextPage"`
	NextStartIndex int  `json:"nextStartindex,omitempty"`
}
