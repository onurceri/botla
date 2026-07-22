> Extracted from: rag-pipeline.md — Section: 2.4 Chunking (`chunker.go`)

## 2.4 Chunking (`chunker.go`)

- [ ] Target chunk size: 512 tokens
- [ ] Overlap: 15% (~77 tokens)
- [ ] Sentence boundary detection (splits on `.`, `!`, `?` with lookahead)
- [ ] Does not split code blocks or tables mid-chunk
- [ ] Handles very short content (returns single chunk)
- [ ] Handles very long content (multiple chunks with overlap)
- [ ] Token counting is consistent with LLM tokenizer
