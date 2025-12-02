package main

import (
	"fmt"
	"strings"

	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/pkg/langconfig"
)

func main() {
	// Ensure config points to correct data path relative to where we run this
	langconfig.Configs["tr"] = langconfig.LanguageConfig{
		Code: "tr",
		Name: "Turkish",
		Abbreviations: []string{
			"Dr.", "Prof.", "vb.", "Av.", "Ecz.", "Doç.", "Yrd.", "Cad.", "Sok.", "Mah.",
		},
		TokenMultiplier: 1.3,
		OCRLanguage:     "tur",
		TokenizerData:   "data/sentences/turkish.json",
	}

	text := "Prof. Dr. Ahmet Bey geldi. Yanında Av. Mehmet de vardı. Bu bir test cümlesidir. Kısaltmalar vb. doğru çalışmalı."
	
	fmt.Println("Original Text:", text)
	fmt.Println("--- Chunks ---")

	chunks, err := rag.ChunkText(text, 50, "tr")
	if err != nil {
		panic(err)
	}

	for i, c := range chunks {
		fmt.Printf("Chunk %d: %s\n", i+1, c.Text)
	}
	
	// Expected: Should not split at "Dr." or "Av." or "vb."
	if len(chunks) > 0 {
		if strings.Contains(chunks[0].Text, "Prof. Dr. Ahmet Bey geldi.") && 
		   strings.Contains(chunks[0].Text, "Yanında Av. Mehmet de vardı.") {
			fmt.Println("\nSUCCESS: Abbreviations handled correctly.")
		} else {
			fmt.Println("\nFAILURE: Incorrect splitting.")
		}
	}
}
