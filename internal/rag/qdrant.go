package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

// VectorClient defines the interface for interacting with a vector database
type VectorClient interface {
	EnsureEmbeddingsCollection(ctx context.Context) error
	UpsertEmbedding(ctx context.Context, id interface{}, vector []float32, payload EmbeddingPayload) error
	SearchSimilar(ctx context.Context, embedding []float32, chatbotID string, topK int) ([]SearchResult, error)
	DeleteBySourceID(ctx context.Context, sourceID string) error
	ScrollChunks(ctx context.Context, sourceID string, limit int, offset interface{}) ([]SearchResult, *string, error)
}

type QdrantClient struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func NewQdrantClientFromEnv() (*QdrantClient, error) {
	u := os.Getenv("QDRANT_URL")
	if u == "" {
		return nil, errors.New("QDRANT_URL is empty")
	}
	k := os.Getenv("QDRANT_API_KEY")
	to := 15 * time.Second
	if v := os.Getenv("QDRANT_TIMEOUT_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			to = time.Duration(n) * time.Millisecond
		}
	}
	return &QdrantClient{baseURL: u, apiKey: k, http: &http.Client{Timeout: to}}, nil
}

type EmbeddingPayload struct {
	ChatbotID    string    `json:"chatbot_id"`
	SourceID     string    `json:"source_id"`
	ChunkIndex   int       `json:"chunk_index"`
	OriginalText string    `json:"original_text"`
	SourceType   string    `json:"source_type"`
	CreatedAt    time.Time `json:"created_at"`
}

type ensureCollectionRequest struct {
	Vectors struct {
		Size     int    `json:"size"`
		Distance string `json:"distance"`
	} `json:"vectors"`
	ShardNumber            int `json:"shard_number"`
	ReplicationFactor      int `json:"replication_factor"`
	WriteConsistencyFactor int `json:"write_consistency_factor"`
}

type qdrantResponse struct {
	Status string          `json:"status"`
	Result json.RawMessage `json:"result"`
}

func (c *QdrantClient) EnsureEmbeddingsCollection(ctx context.Context) error {
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
	body := ensureCollectionRequest{ShardNumber: 2, ReplicationFactor: 1, WriteConsistencyFactor: 1}
	body.Vectors.Size = 1536
	body.Vectors.Distance = "Cosine"
	b, _ := json.Marshal(body)
	req2, _ := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(b))
	setJSONHeaders(req2, c.apiKey)
	res2, err := c.http.Do(req2)
	if err != nil {
		return fmt.Errorf("request to create collection failed: %w", err)
	}
	defer func() { _ = res2.Body.Close() }()
	if res2.StatusCode != http.StatusOK {
		return fmt.Errorf("create collection failed with status %s", res2.Status)
	}
	return nil
}

type upsertPointsRequest struct {
	Points []struct {
		ID      interface{}      `json:"id"`
		Vector  []float32        `json:"vector"`
		Payload EmbeddingPayload `json:"payload"`
	} `json:"points"`
}

func (c *QdrantClient) UpsertEmbedding(ctx context.Context, id interface{}, vector []float32, payload EmbeddingPayload) error {
	reqBody := upsertPointsRequest{Points: make([]struct {
		ID      interface{}      `json:"id"`
		Vector  []float32        `json:"vector"`
		Payload EmbeddingPayload `json:"payload"`
	}, 1)}
	reqBody.Points[0].ID = id
	reqBody.Points[0].Vector = vector
	reqBody.Points[0].Payload = payload
	b, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("%s/collections/embeddings/points?wait=true", c.baseURL)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(b))
	setJSONHeaders(req, c.apiKey)
	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("upsert request failed: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("upsert failed: %s", res.Status)
	}
	return nil
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

type SearchResult struct {
	ID      interface{}      `json:"id"`
	Score   float64          `json:"score"`
	Payload EmbeddingPayload `json:"payload"`
}

func (c *QdrantClient) SearchSimilar(ctx context.Context, embedding []float32, chatbotID string, topK int) ([]SearchResult, error) {
	reqBody := searchRequest{Vector: embedding, Limit: topK, WithPayload: true, Filter: filterBody{Must: []condition{{Key: "chatbot_id", Match: matchBody{Value: chatbotID}}}}}
	b, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("%s/collections/embeddings/points/search", c.baseURL)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	setJSONHeaders(req, c.apiKey)
	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search similar request failed: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed: %s", res.Status)
	}
	var qres qdrantResponse
	if err := json.NewDecoder(res.Body).Decode(&qres); err != nil {
		return nil, fmt.Errorf("decoding qdrant response: %w", err)
	}
	var items []SearchResult
	if err := json.Unmarshal(qres.Result, &items); err != nil {
		return nil, fmt.Errorf("unmarshaling search results: %w", err)
	}
	return items, nil
}

type deleteFilterRequest struct {
	Filter filterBody `json:"filter"`
}

func (c *QdrantClient) DeleteBySourceID(ctx context.Context, sourceID string) error {
	body := deleteFilterRequest{Filter: filterBody{Must: []condition{{Key: "source_id", Match: matchBody{Value: sourceID}}}}}
	b, _ := json.Marshal(body)
	url := fmt.Sprintf("%s/collections/embeddings/points/delete?wait=true", c.baseURL)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	setJSONHeaders(req, c.apiKey)
	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("delete by source id request failed: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("delete failed: %s", res.Status)
	}
	return nil
}

type scrollRequest struct {
	Filter      filterBody  `json:"filter"`
	Limit       int         `json:"limit"`
	WithPayload bool        `json:"with_payload"`
	WithVector  bool        `json:"with_vector"`
	Offset      interface{} `json:"offset,omitempty"`
}

type scrollResult struct {
	Points         []SearchResult `json:"points"`
	NextPageOffset interface{}    `json:"next_page_offset"`
}

// ScrollChunks retrieves paginated chunks for a given sourceID.
// It returns the list of points, the next offset (can be used for next call), and error.
// The nextOffset is nil if there are no more pages.
func (c *QdrantClient) ScrollChunks(ctx context.Context, sourceID string, limit int, offset interface{}) ([]SearchResult, *string, error) {
	reqBody := scrollRequest{
		Filter:      filterBody{Must: []condition{{Key: "source_id", Match: matchBody{Value: sourceID}}}},
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
		return nil, nil, fmt.Errorf("scroll chunks request failed: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("scroll failed: %s", res.Status)
	}

	var qres qdrantResponse
	if err := json.NewDecoder(res.Body).Decode(&qres); err != nil {
		return nil, nil, fmt.Errorf("decoding scroll response: %w", err)
	}

	var result scrollResult
	if err := json.Unmarshal(qres.Result, &result); err != nil {
		return nil, nil, fmt.Errorf("unmarshaling scroll result: %w", err)
	}

	var nextOffset *string
	if result.NextPageOffset != nil {
		// handle UUID (string) or int offset
		switch v := result.NextPageOffset.(type) {
		case string:
			nextOffset = &v
		case float64: // json unmarshals numbers as float64
			s := fmt.Sprintf("%.0f", v)
			nextOffset = &s
		default:
			// if it's something else, try string conversion
			s := fmt.Sprintf("%v", v)
			nextOffset = &s
		}
	}

	return result.Points, nextOffset, nil
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
