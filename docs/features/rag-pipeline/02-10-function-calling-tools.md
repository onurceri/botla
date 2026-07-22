> Extracted from: rag-pipeline.md — Section: 2.10 Function Calling / Tools

## 2.10 Function Calling / Tools (`tools.go`, `tool_name_generator.go`, `tool_executor.go`)

- [ ] Tool definitions are generated dynamically from chatbot capabilities
- [ ] Each tool has a unique name (no collisions)
- [ ] Tool execution is sandboxed
- [ ] Tool results are incorporated into the LLM response
- [ ] Tool execution errors are handled gracefully
