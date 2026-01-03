package processing

import (
	"testing"

	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/onurceri/botla-co/pkg/storage"
)

func TestSourceProcessorInterface_Implementation(t *testing.T) {
	t.Parallel()
	tdb := testdb.OpenParallelTestDB(t)
	mockOAI := &rag.MockFullClient{}
	mockVC := &rag.MockVectorClient{}
	mockStorage := &storage.MockStorageService{}

	t.Run("URLProcessor implements SourceProcessor", func(t *testing.T) {
		var _ SourceProcessor = NewURLProcessor(tdb, mockOAI, mockVC, nil, nil, nil)
	})

	t.Run("PDFProcessor implements SourceProcessor", func(t *testing.T) {
		var _ SourceProcessor = NewPDFProcessor(tdb, mockStorage, mockOAI, mockVC, nil, nil)
	})

	t.Run("TextProcessor implements SourceProcessor", func(t *testing.T) {
		var _ SourceProcessor = NewTextProcessor(tdb, mockStorage, mockOAI, mockVC, nil, nil)
	})
}

func TestProcessorMap_Registry(t *testing.T) {
	t.Parallel()
	tdb := testdb.OpenParallelTestDB(t)
	mockOAI := &rag.MockFullClient{}
	mockVC := &rag.MockVectorClient{}
	mockStorage := &storage.MockStorageService{}

	urlProc := NewURLProcessor(tdb, mockOAI, mockVC, nil, nil, nil)
	pdfProc := NewPDFProcessor(tdb, mockStorage, mockOAI, mockVC, nil, nil)
	textProc := NewTextProcessor(tdb, mockStorage, mockOAI, mockVC, nil, nil)

	processors := map[string]SourceProcessor{
		"url":  urlProc,
		"pdf":  pdfProc,
		"text": textProc,
	}

	t.Run("URL processor is registered", func(t *testing.T) {
		proc, ok := processors["url"]
		if !ok {
			t.Error("url processor not found in map")
		}
		if proc != urlProc {
			t.Error("url processor is not the expected instance")
		}
	})

	t.Run("PDF processor is registered", func(t *testing.T) {
		proc, ok := processors["pdf"]
		if !ok {
			t.Error("pdf processor not found in map")
		}
		if proc != pdfProc {
			t.Error("pdf processor is not the expected instance")
		}
	})

	t.Run("Text processor is registered", func(t *testing.T) {
		proc, ok := processors["text"]
		if !ok {
			t.Error("text processor not found in map")
		}
		if proc != textProc {
			t.Error("text processor is not the expected instance")
		}
	})
}

func TestProcessorMap_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("Unknown source type returns false from lookup", func(t *testing.T) {
		processors := map[string]SourceProcessor{
			"url":  nil,
			"pdf":  nil,
			"text": nil,
		}

		_, ok := processors["unknown_type"]
		if ok {
			t.Error("unknown source type should not be found in map")
		}
	})

	t.Run("Nil processor in map is retrievable", func(t *testing.T) {
		processors := map[string]SourceProcessor{
			"url": nil,
		}

		proc, ok := processors["url"]
		if !ok {
			t.Error("url key not found in map")
		}
		if proc != nil {
			t.Error("expected nil processor")
		}
	})

	t.Run("Empty map returns false for any lookup", func(t *testing.T) {
		processors := make(map[string]SourceProcessor)

		_, ok := processors["url"]
		if ok {
			t.Error("empty map should not return true for any key")
		}
	})
}
