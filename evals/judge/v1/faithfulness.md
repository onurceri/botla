RUBRIC: FAITHFULNESS (groundedness)

Decide whether every factual claim in the ANSWER is supported by the CONTEXT that the chatbot was given. This detects hallucination. It does not matter whether the answer is complete or whether it answers the question well.

Procedure:
1. Split the answer into atomic factual claims (one fact each: a number, a date, a plan feature, a procedure step, a yes/no statement about the product, etc.).
2. Do NOT list as claims: greetings, offers to help, statements that the assistant lacks information, hedges, restatements of the user's question, or generic advice that contains no product-specific facts.
3. For each claim, mark supported=true only if the context states it or it follows by simple arithmetic or direct logical combination of context statements. Otherwise supported=false. A claim that contradicts the context is unsupported. A claim about something the context never mentions is unsupported.
4. If the context is empty, any product-specific factual claim is unsupported.

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
{"claims": [{"claim": "<atomic claim>", "supported": true, "evidence": "<short quote from context, or empty>"}], "rationale": "<one or two sentences>"}
Use an empty "claims" array if the answer makes no factual claims.
