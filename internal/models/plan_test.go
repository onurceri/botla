package models

import (
    "encoding/json"
    "testing"
)

func TestPlanConfig_ValueScan(t *testing.T) {
    pc := PlanConfig{
        Scraping: ScrapingConfig{DynamicEnabled: true, MaxURLsPerBot: 3, MaxPagesPerCrawl: 2},
        Files:    FilesConfig{OCREnabled: false, MaxSizeMB: 10, MaxFilesPerBot: 5, MaxFilesTotal: 50, TotalStorageMB: 100},
        Chat:     ChatConfig{AllowedModels: []string{"gpt-4o-mini"}, MaxMonthlyTokens: 1000, RAG: RAGConfig{TopK: 3, MaxContextTokens: 512}},
        MaxMonthlyIngestions:     5,
        MaxMonthlyEmbeddingTokens: 10000,
        MinReAddCooldownMinutes:   30,
    }
    v, err := pc.Value()
    if err != nil { t.Fatalf("value: %v", err) }
    var got PlanConfig
    if err := got.Scan(v); err != nil { t.Fatalf("scan: %v", err) }
    // roundtrip check
    b1, _ := json.Marshal(pc)
    b2, _ := json.Marshal(got)
    if string(b1) != string(b2) { t.Fatalf("mismatch: %s vs %s", string(b1), string(b2)) }
}

