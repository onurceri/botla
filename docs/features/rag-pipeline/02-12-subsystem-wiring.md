> Extracted from: rag-pipeline.md — Section: 2.12 Subsystem Wiring (`subsystem.go`)

## 2.12 Subsystem Wiring (`subsystem.go`)

- [ ] All RAG components are wired together in a subsystem struct
- [ ] Dependencies are injected (no global state)
- [ ] Health check endpoint reflects subsystem state
- [ ] Graceful shutdown drains in-flight requests
