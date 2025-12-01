package main

import (
    "context"
    "fmt"
    "os"
    "time"
    "github.com/onurceri/botla-co/internal/rag"
)

func main() {
    c, err := rag.NewQdrantClientFromEnv()
    if err != nil {
        fmt.Println("error:", err)
        os.Exit(1)
    }
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    if err := c.EnsureEmbeddingsCollection(ctx); err != nil {
        fmt.Println("ensure collection failed:", err)
        os.Exit(1)
    }
    fmt.Println("embeddings collection is ready")
}

