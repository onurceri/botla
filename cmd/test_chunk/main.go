package main

import (
	"fmt"
	"strings"

	"github.com/onurceri/botla-co/internal/processing"
)

func main() {
	// Generate a long string (approx 6000 chars)
	longText := strings.Repeat("This is a test sentence to simulate content. ", 150) // 45 chars * 150 = 6750 chars
	
	fmt.Printf("Input length: %d\n", len(longText))

	chunks := processing.ChunkText(longText, 1500, 200)
	fmt.Printf("Chunk count: %d\n", len(chunks))
	
	for i, c := range chunks {
		fmt.Printf("Chunk %d length: %d\n", i, len(c))
	}
}
