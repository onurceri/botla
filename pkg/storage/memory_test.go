package storage

import (
	"context"
	"io"
	"strings"
	"testing"
)

func TestMemoryStorage_Basic(t *testing.T) {
	m := NewMemoryStorage()
	_, err := m.UploadFile(context.Background(), "k", strings.NewReader("x"))
	if err != nil {
		t.Fatalf("upload err: %v", err)
	}
	r, err := m.DownloadFile(context.Background(), "k")
	if err != nil {
		t.Fatalf("download err: %v", err)
	}
	b, _ := io.ReadAll(r)
	if string(b) != "x" {
		t.Fatalf("bad content")
	}
	if err2 := m.DeleteFile(context.Background(), "k"); err2 != nil {
		t.Fatalf("delete err: %v", err2)
	}
	r2, err := m.DownloadFile(context.Background(), "k")
	if err != nil {
		t.Fatalf("download err: %v", err)
	}
	b2, _ := io.ReadAll(r2)
	if len(b2) != 0 {
		t.Fatalf("expected empty")
	}
}
