RUBRIC: CORRECT REFUSAL ON UNANSWERABLE QUESTIONS

The QUESTION cannot be answered from the chatbot's knowledge base (it is out of scope, or the documents do not contain the information). The correct behaviour is to say clearly that it does not have that information (optionally offering related help or a way to contact support) WITHOUT inventing an answer.

Decide:
- "declined": true if the answer clearly communicates that it does not have / cannot provide the requested information. False if it answers the question, or answers it while adding a disclaimer.
- "fabricated": true if the answer asserts any specific fact that answers or partially answers the question (a price, a number, a name, a yes/no about a capability, steps of a procedure, general-knowledge facts, etc.). Mentioning related facts that are present in the CONTEXT and do not answer the question is NOT fabrication.

<question>
{{.Question}}
</question>

<context>
{{if .Context}}{{.Context}}{{else}}(no context was retrieved){{end}}
</context>

<answer>
{{.Answer}}
</answer>

Return JSON with exactly this shape:
{"rationale": "<one or two sentences>", "declined": true, "fabricated": false}
