package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/onurceri/botla-co/internal/db"
)

func main() {
	dsn := "postgres://botla:botla@localhost:5432/botla_dev?sslmode=disable"
	pool, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer pool.Close()

	chatbotID := "86b8649b-032d-4871-9973-56c9d3122bed"
	ctx := context.Background()

	fmt.Printf("Fixing suggestions for chatbot %s...\n", chatbotID)

	// Copy-paste of the fixed logic since I cannot easily import internal/processing if it has other dependencies I don't want to init
	// Actually I can import internal/db.
	// But the logic is in internal/processing/sources_queue.go which is not exported as a standalone function easily without SourceQueue struct?
	// Wait, aggregateAndPersistChatbotSuggestions IS a standalone function in sources_queue.go but it is NOT exported (lowercase 'a').
	// So I have to duplicate the logic here.

	// Fetch existing chatbot suggestions
	var existing []byte
	var currentSuggestions []string
	_ = pool.QueryRowContext(ctx, `SELECT suggested_questions FROM chatbots WHERE id=$1`, chatbotID).Scan(&existing)

	if len(existing) > 0 {
		_ = json.Unmarshal(existing, &currentSuggestions)
	}

	// Initialize set with existing suggestions to avoid duplicates
	uniq := map[string]struct{}{}
	out := make([]string, 0, 6)

	for _, q := range currentSuggestions {
		t := strings.TrimSpace(q)
		if t == "" { continue }
		if len(t) > 120 { t = t[:120] }
		k := strings.ToLower(t)
		if _, ok := uniq[k]; ok { continue }
		uniq[k] = struct{}{}
		out = append(out, t)
	}

	// If we already have 6 or more questions, we don't need to fetch more
	if len(out) >= 6 {
		fmt.Println("Already have 6 or more suggestions.")
		return
	}

	rows, err := pool.QueryContext(ctx, `SELECT suggested_questions FROM data_sources WHERE chatbot_id=$1 AND suggested_questions IS NOT NULL`, chatbotID)
	if err != nil {
		log.Fatalf("query sources failed: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var arr []byte
		if err := rows.Scan(&arr); err != nil { continue }
		var items []string
		if err := json.Unmarshal(arr, &items); err != nil { continue }
		for _, it := range items {
			t := strings.TrimSpace(it)
			if t == "" { continue }
			if len(t) > 120 { t = t[:120] }
			k := strings.ToLower(t)
			if _, ok := uniq[k]; ok { continue }
			uniq[k] = struct{}{}
			out = append(out, t)
			if len(out) >= 6 { break }
		}
		if len(out) >= 6 { break }
	}

	fmt.Printf("Updating with %d suggestions: %v\n", len(out), out)
	err = db.UpdateChatbotSuggestions(ctx, pool, chatbotID, out)
	if err != nil {
		log.Fatalf("update failed: %v", err)
	}
	fmt.Println("Success!")
}
