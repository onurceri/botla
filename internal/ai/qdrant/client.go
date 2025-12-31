package qdrant

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/onurceri/botla-co/internal/ai"
)

// Config holds configuration for Qdrant client
type Config struct {
	URL     string        // Required
	APIKey  string        // Optional
	Timeout time.Duration // Optional, defaults to 15s
}

// Client implements ai.VectorStore for Qdrant
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// Verify interface compliance at compile time
var _ ai.VectorStore = (*Client)(nil)

// New creates a new Qdrant client with the given configuration
func New(config Config, client *http.Client) (*Client, error) {
	if config.URL == "" {
		return nil, errors.New("url is required")
	}

	if config.Timeout == 0 {
		config.Timeout = 15 * time.Second
	}
	if client == nil {
		client = &http.Client{Timeout: config.Timeout}
	}
	return &Client{
		baseURL: config.URL,
		apiKey:  config.APIKey,
		http:    client,
	}, nil
}

// EnsureCollection ensures the embeddings collection exists
func (c *Client) EnsureCollection(ctx context.Context) error {
	url := fmt.Sprintf("%s/collections/embeddings", c.baseURL)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	setHeaders(req, c.apiKey)
	res, err := c.http.Do(req)
	if err == nil && res.StatusCode == http.StatusOK {
		_ = res.Body.Close()
		return nil
	}
	if res != nil {
		_ = res.Body.Close()
	}

	// Create collection with default settings
	body := ensureCollectionRequest{
		ShardNumber:            2,
		ReplicationFactor:      1,
		WriteConsistencyFactor: 1,
	}
	body.Vectors.Size = 1536
	body.Vectors.Distance = "Cosine"

	b, _ := json.Marshal(body)
	req2, _ := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(b))
	setJSONHeaders(req2, c.apiKey)
	res2, err := c.http.Do(req2)
	if err != nil {
		return fmt.Errorf("http do: %w", err)
	}
	defer func() { _ = res2.Body.Close() }()
	if res2.StatusCode != http.StatusOK {
		return fmt.Errorf("create collection failed: %s", res2.Status)
	}
	return nil
}

// Upsert inserts or updates a vector with associated metadata
func (c *Client) Upsert(ctx context.Context, id interface{}, vector []float32, payload ai.VectorPayload) error {
	reqBody := upsertPointsRequest{Points: make([]point, 1)}
	reqBody.Points[0].ID = id
	reqBody.Points[0].Vector = vector
	reqBody.Points[0].Payload = payload

	b, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("%s/collections/embeddings/points?wait=true", c.baseURL)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(b))
	setJSONHeaders(req, c.apiKey)
	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("http do upsert: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("upsert failed: %s", res.Status)
	}
	return nil
}

// Search performs similarity search with filters
func (c *Client) Search(ctx context.Context, vector []float32, filter ai.SearchFilter, limit int) ([]ai.SearchResult, error) {
	var conditions []condition
	if filter.ChatbotID != "" {
		conditions = append(conditions, condition{
			Key:   "chatbot_id",
			Match: matchBody{Value: filter.ChatbotID},
		})
	}
	if filter.SourceID != "" {
		conditions = append(conditions, condition{
			Key:   "source_id",
			Match: matchBody{Value: filter.SourceID},
		})
	}

	reqBody := searchRequest{
		Vector:      vector,
		Limit:       limit,
		WithPayload: true,
		Filter:      filterBody{Must: conditions},
	}

	b, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("%s/collections/embeddings/points/search", c.baseURL)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	setJSONHeaders(req, c.apiKey)
	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http do search: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed: %s", res.Status)
	}

	var qres qdrantResponse
	if err := json.NewDecoder(res.Body).Decode(&qres); err != nil {
		return nil, fmt.Errorf("decode qdrant response: %w", err)
	}

	var items []ai.SearchResult
	if err := json.Unmarshal(qres.Result, &items); err != nil {
		return nil, fmt.Errorf("unmarshal qdrant result: %w", err)
	}
	return items, nil
}

// Delete removes vectors matching the filter
func (c *Client) Delete(ctx context.Context, filter ai.DeleteFilter) error {
	body := deleteFilterRequest{
		Filter: filterBody{
			Must: []condition{
				{Key: "source_id", Match: matchBody{Value: filter.SourceID}},
			},
		},
	}

	b, _ := json.Marshal(body)
	url := fmt.Sprintf("%s/collections/embeddings/points/delete?wait=true", c.baseURL)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	setJSONHeaders(req, c.apiKey)
	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("http do delete: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("delete failed: %s", res.Status)
	}
	return nil
}

// Scroll retrieves vectors in pages with optional offset for pagination
func (c *Client) Scroll(ctx context.Context, filter ai.SearchFilter, limit int, offset interface{}) ([]ai.SearchResult, interface{}, error) {
	var conditions []condition
	if filter.ChatbotID != "" {
		conditions = append(conditions, condition{
			Key:   "chatbot_id",
			Match: matchBody{Value: filter.ChatbotID},
		})
	}
	if filter.SourceID != "" {
		conditions = append(conditions, condition{
			Key:   "source_id",
			Match: matchBody{Value: filter.SourceID},
		})
	}

	reqBody := scrollRequest{
		Filter:      filterBody{Must: conditions},
		Limit:       limit,
		WithPayload: true,
		WithVector:  false,
		Offset:      offset,
	}

	b, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("%s/collections/embeddings/points/scroll", c.baseURL)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	setJSONHeaders(req, c.apiKey)

	res, err := c.http.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("http do scroll: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("scroll failed: %s", res.Status)
	}

	var qres qdrantResponse
	if err := json.NewDecoder(res.Body).Decode(&qres); err != nil {
		return nil, nil, fmt.Errorf("decode qdrant scroll response: %w", err)
	}

	var result scrollResult
	if err := json.Unmarshal(qres.Result, &result); err != nil {
		return nil, nil, fmt.Errorf("unmarshal qdrant scroll result: %w", err)
	}

	return result.Points, result.NextPageOffset, nil
}

// Internal types for Qdrant API

type ensureCollectionRequest struct {
	Vectors struct {
		Size     int    `json:"size"`
		Distance string `json:"distance"`
	} `json:"vectors"`
	ShardNumber            int `json:"shard_number"`
	ReplicationFactor      int `json:"replication_factor"`
	WriteConsistencyFactor int `json:"write_consistency_factor"`
}

type point struct {
	ID      interface{}      `json:"id"`
	Vector  []float32        `json:"vector"`
	Payload ai.VectorPayload `json:"payload"`
}

type upsertPointsRequest struct {
	Points []point `json:"points"`
}

type searchRequest struct {
	Vector      []float32  `json:"vector"`
	Limit       int        `json:"limit"`
	WithPayload bool       `json:"with_payload"`
	Filter      filterBody `json:"filter"`
}

type filterBody struct {
	Must []condition `json:"must"`
}

type condition struct {
	Key   string    `json:"key"`
	Match matchBody `json:"match"`
}

type matchBody struct {
	Value string `json:"value"`
}

type deleteFilterRequest struct {
	Filter filterBody `json:"filter"`
}

type scrollRequest struct {
	Filter      filterBody  `json:"filter"`
	Limit       int         `json:"limit"`
	WithPayload bool        `json:"with_payload"`
	WithVector  bool        `json:"with_vector"`
	Offset      interface{} `json:"offset,omitempty"`
}

type scrollResult struct {
	Points         []ai.SearchResult `json:"points"`
	NextPageOffset interface{}       `json:"next_page_offset"`
}

type qdrantResponse struct {
	Status string          `json:"status"`
	Result json.RawMessage `json:"result"`
}

func setHeaders(r *http.Request, apiKey string) {
	if apiKey != "" {
		r.Header.Set("api-key", apiKey)
	}
}

func setJSONHeaders(r *http.Request, apiKey string) {
	r.Header.Set("Content-Type", "application/json")
	setHeaders(r, apiKey)
}
