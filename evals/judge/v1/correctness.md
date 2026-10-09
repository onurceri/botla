RUBRIC: ANSWER CORRECTNESS

Decide how well the ANSWER covers each EXPECTED FACT for the QUESTION. The expected facts are the reference answer written by a human.

For every expected fact, assign exactly one status:
- "covered": the answer states the fact (wording may differ; numbers, units, plan names and conditions must match).
- "partial": the answer states part of the fact or states it vaguely (for example "a couple of years" instead of "24 months").
- "missing": the answer does not address the fact.
- "contradicted": the answer states something incompatible with the fact.

Additional information in the answer does not change the status of a fact, unless it contradicts the fact. Do not judge style or length.

<question>
{{.Question}}
</question>

<expected_facts>
{{range $i, $f := .ExpectedFacts}}{{inc $i}}. {{$f}}
{{end}}</expected_facts>

<answer>
{{.Answer}}
</answer>

Return JSON with exactly this shape, one entry per expected fact, in order:
{"facts": [{"fact": "<expected fact>", "status": "covered|partial|missing|contradicted", "note": "<short justification>"}], "rationale": "<one or two sentences>"}
