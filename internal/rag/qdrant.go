package rag

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "os"
    "time"
)

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
    return &QdrantClient{baseURL: u, apiKey: k, http: &http.Client{Timeout: 15 * time.Second}}, nil
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
    ShardNumber          int `json:"shard_number"`
    ReplicationFactor    int `json:"replication_factor"`
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
        res.Body.Close()
        return nil
    }
    if res != nil {
        res.Body.Close()
    }
    body := ensureCollectionRequest{ShardNumber: 2, ReplicationFactor: 1, WriteConsistencyFactor: 1}
    body.Vectors.Size = 1536
    body.Vectors.Distance = "Cosine"
    b, _ := json.Marshal(body)
    req2, _ := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(b))
    setJSONHeaders(req2, c.apiKey)
    res2, err := c.http.Do(req2)
    if err != nil {
        return err
    }
    defer res2.Body.Close()
    if res2.StatusCode != http.StatusOK {
        return fmt.Errorf("create collection failed: %s", res2.Status)
    }
    return nil
}

type upsertPointsRequest struct {
    Points []struct {
        ID      interface{}     `json:"id"`
        Vector  []float32       `json:"vector"`
        Payload EmbeddingPayload `json:"payload"`
    } `json:"points"`
}

func (c *QdrantClient) UpsertEmbedding(ctx context.Context, id interface{}, vector []float32, payload EmbeddingPayload) error {
    reqBody := upsertPointsRequest{Points: make([]struct {
        ID      interface{}     `json:"id"`
        Vector  []float32       `json:"vector"`
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
        return err
    }
    defer res.Body.Close()
    if res.StatusCode != http.StatusOK {
        return fmt.Errorf("upsert failed: %s", res.Status)
    }
    return nil
}

type searchRequest struct {
    Vector       []float32   `json:"vector"`
    Limit        int         `json:"limit"`
    WithPayload  bool        `json:"with_payload"`
    Filter       filterBody  `json:"filter"`
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
    ID      interface{}     `json:"id"`
    Score   float64         `json:"score"`
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
        return nil, err
    }
    defer res.Body.Close()
    if res.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("search failed: %s", res.Status)
    }
    var qres qdrantResponse
    if err := json.NewDecoder(res.Body).Decode(&qres); err != nil {
        return nil, err
    }
    var items []SearchResult
    if err := json.Unmarshal(qres.Result, &items); err != nil {
        return nil, err
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
        return err
    }
    defer res.Body.Close()
    if res.StatusCode != http.StatusOK {
        return fmt.Errorf("delete failed: %s", res.Status)
    }
    return nil
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
