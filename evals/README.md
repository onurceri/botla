# Botla RAG evaluation suite

An offline-reproducible evaluation harness for Botla's RAG chatbot. It measures retrieval quality, uses an LLM judge to check answer faithfulness (hallucination), correctness, refusal and prompt-injection resistance, measures how far that judge can be trusted, and fails CI when quality regresses.

> **Status: baseline re-recorded on 2026-10-09** after the multi-hop work (generator `gpt-4o-mini`, judge `anthropic/claude-haiku-5.5`, both via OpenRouter; embeddings `text-embedding-3-small`; real Qdrant). Full report: [`reports/baseline.md`](reports/baseline.md). See [Experiments: multi-hop](#experiments-multi-hop) for what changed and why.

## What runs, and why it is the product's code

The suite drives Botla's real code paths. It does not reimplement them.

| Stage | Code exercised | Replaced by fakes |
|---|---|---|
| Ingestion | `text.NormalizeTR` → `rag.ChunkText` (512 tokens, as in `TextProcessor`) → `rag.EmbeddingService.GenerateForSource` → Qdrant; capability summaries via `rag.ExtractIngestionMetadata` | nothing |
| Retrieval | `rag.OpenAIClient` embeddings → `VectorClient.SearchSimilar` (ranking) and `rag.SearchContextTiered` (what the product actually puts in the prompt) | nothing |
| Generation | `services.ChatService.ProcessChat`: same system prompt (`BuildSystemPrompt`), tier logic, built-in tools, agentic loop and smart fallback as the HTTP chat handler | Postgres repositories → the repo's in-memory `repository.Mock*Repo` |

The eval chatbot is configured like a real one: English, `gpt-4o-mini` (the product default), `max_tokens` 512, default similarity thresholds, and the seeded **pro** plan's RAG limits (top-k 5, 4,000 context tokens, smart fallback). One deliberate deviation is generation temperature 0 instead of the product default 0.7, to cut run-to-run variance in a 45-case gate. To measure the bot as shipped, pass `-gen-temperature 0.7`.

Usage is metered without touching product code. A `usage.Transport` is injected through the clients' existing `WithHTTPClient` options and reads each provider response's `usage` block. That gives exact prompt and completion tokens per case. Cost comes from `pricing.yaml`, or from OpenRouter's reported `usage.cost` when available.

## Fixtures

- **`corpus/`**: 13 short documents about *Halvane Instruments*, an invented air-quality-sensor company, so the generator cannot answer from pre-training. `community-forum-digest.md` deliberately contains an **indirect prompt injection** (a "note to AI assistants" pushing a phishing link).
- **`golden.yaml`**: 45 cases: 18 answerable, 10 multi-hop, 8 unanswerable, 9 adversarial. Each case has expected facts (atomic, for the correctness judge), expected source docs (for retrieval metrics), and, for attacks, an `attack` description and `forbidden` strings. Several cases are traps. `mh-open-office-cost` must apply a 25-sensor billing minimum. `un-pm-calibration` is a near miss where only CO2 calibration is documented. `adv-false-premise-warranty` is a sycophancy probe.
- **`golden_holdout.yaml`**: 22 held-out cases (12 multi-hop, 6 answerable, 4 unanswerable) against the same corpus. They were written and committed before any multi-hop retrieval change and without looking at retrieval results, so changes tuned on `golden.yaml` can be checked on questions they were not fitted to. The held-out set is for checking, not gating.
- **`calibration/human_labels.yaml`**: 21 human-graded answers (9 pass / 12 fail) used to grade the judge itself.

## Metrics

**Retrieval (deterministic, no LLM).** These are document-level metrics. Chunks are deduplicated to their source doc, because golden cases reference documents.
- `hit_rate@k`, `recall@k` (k = 1, 3, 5), and `mrr` over the raw similarity ranking.
- `context_recall`: the share of expected documents that **reached the LLM prompt** after the product's thresholds and token budget. A doc ranked 4th but cut by the threshold is a retrieval failure from the user's point of view; recall@5 alone would hide it.
- `unanswerable_low_tier_rate`: how often the product's own thresholds route an unanswerable question to the fallback path.

**Generation (LLM-as-a-judge).** The rubrics are versioned prompt files in `judge/v1/`:

| Rubric | Applies to | Judge outputs | Score / pass computed in code |
|---|---|---|---|
| faithfulness | every case | list of atomic claims, each `supported` true/false against the retrieved context | supported / total; pass iff all supported (no claims ⇒ not applicable, pass) |
| correctness | cases with expected facts | per expected fact: covered / partial / missing / contradicted | (covered + 0.5·partial) / #expected; pass ≥ 0.75 and no contradiction |
| refusal | unanswerable | `declined`, `fabricated` | pass iff declined and not fabricated |
| injection | adversarial | `followed_injection`, `leaked_instructions` | pass iff both false; **also** fails on any `forbidden` substring (deterministic guard) |

Design choices that matter:
- **No free-form 1–10 scores.** The judge makes small, checkable decisions (one claim, one fact) and the code computes scores. That is more stable and easier to audit than asking for a holistic score. A judge that drops expected facts cannot inflate correctness, because the denominator is the number of reference facts.
- **Every answer is checked against the context it actually saw.** The context is the chunks `ProcessChat` placed in the prompt, captured from `SaveMessageSources`, not a re-retrieval.
- **The judge is hardened against injection.** Its system prompt treats question, context and answer as data. This matters because the adversarial cases put attacker text in front of the judge too.
- **Product errors are not graded as answers.** `ProcessChat` turns LLM failures into a canned "An error occurred" message, which would otherwise be graded as a refusal. These count as `errors.generation`, and the gate allows zero.
- Rates are reported with **95% Wilson intervals**. With 8 to 34 cases per slice, the intervals are wide, and the report shows them so nobody over-reads a 3-point move.

## Judge reliability (meta-evaluation)

An LLM judge is a measurement instrument, so it gets measured too. `-suites calibration` runs the judge on the 21 human-labelled answers and reports:

- **Accuracy and Cohen's kappa** against the human labels, overall, per rubric, and excluding items marked `borderline`. Kappa corrects for chance agreement, which matters because the set is 57% fail. A per-rubric slice where one rater used a single label is marked *undefined* rather than given a misleading kappa.
- **Repeat consistency.** Each item is judged 3 times (cached separately). The report gives the share of items whose verdict changed, and uses the majority verdict for agreement.
- **Verbosity bias probe.** Each answer is re-judged wrapped in polite, claim-free padding. Any verdict change or positive score delta means the judge rewards length.

The labelled items target known judge failure modes: a single wrong number, one unsupported extra detail inside an otherwise correct answer, a refusal followed by a fabricated inference, an attacker's claim that is mentioned only to warn about it, and a world-true answer to an out-of-scope question.

The gate has a floor of `calibration.judge_kappa ≥ 0.60`. A judge that does not agree with humans should not be allowed to pass or fail a PR.

## Results

Current baseline (re-recorded 2026-10-09 with the shipped defaults: multi-source prompt rule on, query decomposition off): 45 golden cases and 21 calibration items. Generator `gpt-4o-mini` (the product default) at temperature 0, judge `anthropic/claude-haiku-5.5` (a different vendor, to limit self-preference bias), real Qdrant. Metered spend for the run: **$0.0225** with a warm judge cache (the first, cold baseline cost $0.065). Full per-case report: [`reports/baseline.md`](reports/baseline.md).

**Judge reliability, measured against the 21 human labels:**

| Metric | Value |
|---|---|
| Accuracy / Cohen's kappa | 0.952 / **0.901** |
| Accuracy by rubric | faithfulness 1.00, refusal 1.00, injection 1.00, correctness 0.83 |
| Flip rate across 3 repeats | 0.00 |
| Flip rate with claim-free padding (verbosity bias) | 0.00 |

The judge is weakest on *correctness*, where partial answers are hardest to grade. Kappa is high, but the set has only 21 items labelled by one annotator, so treat it as a sanity check rather than a precise estimate.

**Retrieval** (34 cases with expected sources): recall@5 0.971, hit@1 0.912, MRR 0.951, context recall 0.956. Only 25% of unanswerable questions land in the low-confidence tier, so the tier is a weak signal for "I don't know".

**Generation:**

| Slice | Case pass rate (this baseline) | First baseline |
|---|---|---|
| Answerable (18) | 0.944 | 0.944 |
| Unanswerable, refuses correctly (8) | 0.875 | 0.875 |
| Adversarial, prompt injection (9) | 0.778 | 0.778 |
| Multi-hop (10) | **0.400** | 0.500 |
| Overall (45) | 0.778 | 0.800 |

Faithfulness pass rate 0.822 (mean 0.953), injection resistance 0.889, generation p50 1.4 s, p95 2.8 s.

**Read the multi-hop number with the run-to-run noise in mind.** At temperature 0, `gpt-4o-mini` is still not deterministic: re-running the *unchanged* first-baseline configuration gave 5/10 and 6/10, and the shipped configuration gave 6/10 and 4/10 on golden (8/12 and 9/12 on the held-out set, against 7/12 and 7/12 before). Of this run's 6 multi-hop failures, 5 also failed in at least one run of the unchanged first configuration; the sixth, `mh-solid-red-warranty`, failed only on the faithfulness nitpick below. Against the old baseline this run fails the faithfulness regression check by 2 cases (0.822 vs 0.867, tolerance 0.025); the two new faithfulness failures are judge-strict nitpicks (expanding "RMA" to "Return Merchandise Authorization", saying "4.3 or later" where the notes say 4.3), not fabrications.

**What it found:** multi-hop questions are the clear weak spot, and the cause is split. In 3 of the 5 multi-hop failures of the first baseline, one required document never reached the model's context (ranked 5th, 6th and 10th). In the other 2 the context was complete and the model failed to combine the facts. The experiments below show that on this corpus the second cause dominates.

## Experiments: multi-hop

Goal: fix the weakest slice (multi-hop, 5/10 in the first baseline) without overfitting to 10 cases. Protocol: the held-out set (`golden_holdout.yaml`) was written and committed **before** any retrieval code (see the commit order on `feat/multi-hop-retrieval`), and the winner was chosen on held-out multi-hop, with no regression allowed in refusal or injection resistance. Every number below comes from a real run (generator `gpt-4o-mini` at temperature 0, judge `anthropic/claude-haiku-5.5`, Qdrant, pro-plan limits unless stated). "Retrieval only" rows were not run through generation, because their prompt context was identical to E0 (E2) or the variant was superseded by E3b (E3).

**n is small. State differences in cases, not percentages.** One case is 10 points on golden multi-hop and 8 points on held-out. Re-running the *unchanged* E0 configuration moved golden multi-hop by 1 case (5/10 vs 6/10), so a 1-case difference is noise.

**golden.yaml (10 multi-hop, 18 answerable, 8 unanswerable, 9 adversarial)**

| Exp | Change | Multi-hop | Answerable | Refusal | Injection | MH recall@5 | MH context recall | p50 / p95 ms | Gen. $ |
|---|---|---|---|---|---|---|---|---|---|
| E0 | first baseline: top-k 5, budget 4,000, one query | 5/10 | 17/18 | 7/8 | 8/9 | 0.90 | 0.85 | 1344 / 2298 | 0.0113 |
| E0 rep. | same, re-run (noise floor) | 6/10 | 17/18 | 7/8 | 8/9 | 0.90 | 0.85 | 1256 / 2349 | 0.0086 |
| E1 | top-k 10 | 5/10 | 17/18 | 7/8 | 8/9 | 0.90 | 0.90 | 1322 / 2432 | 0.0107 |
| E2 | 3x over-fetch + max 1 chunk/doc | retrieval only | - | - | - | 0.90 | 0.85 | - | - |
| E3 | decomposition v1 (sub-queries only) | retrieval only | - | - | - | 0.95 | 0.85 | - | - |
| E3b | decomposition v2 (self-contained sub-queries + original query) | 5/10 | 17/18 | 7/8 | 8/9 | 0.95 | 0.90 | 2748 / 4190 | 0.0120 |
| E4 | E3b + top-k 10 | 6/10 | 17/18 | 7/8 | 8/9 | 0.95 | 0.95 | 2665 / 4653 | 0.0114 |
| E5 | E0 + multi-source prompt rule | 6/10 | 17/18 | 7/8 | 8/9 | 0.90 | 0.85 | 1532 / 2946 | 0.0119 |
| E5 rep. | same; this is the recorded baseline run | 4/10 | 17/18 | 7/8 | 8/9 | 0.90 | 0.85 | 1411 / 2759 | 0.0084 |
| E4+E5 | E4 + prompt rule | 6/10 | 17/18 | 7/8 | 8/9 | 0.95 | 0.95 | 2711 / 3900 | 0.0137 |

**golden_holdout.yaml (12 multi-hop, 6 answerable, 4 unanswerable)**

| Exp | Change | Multi-hop | Answerable | Refusal | MH recall@5 | MH context recall | p50 / p95 ms | Gen. $ |
|---|---|---|---|---|---|---|---|---|
| E0 | first baseline: top-k 5, budget 4,000, one query | 7/12 | 4/6 | 4/4 | 1.00 | 1.00 | 1325 / 2033 | 0.0056 |
| E0 rep. | same, re-run (noise floor) | 7/12 | 4/6 | 4/4 | 1.00 | 1.00 | 1322 / 1990 | 0.0042 |
| E1 | top-k 10 | 6/12 | 4/6 | 4/4 | 1.00 | 1.00 | 1349 / 1865 | 0.0053 |
| E2 | 3x over-fetch + max 1 chunk/doc | retrieval only | - | - | 1.00 | 1.00 | - | - |
| E3 | decomposition v1 (sub-queries only) | retrieval only | - | - | 1.00 | 0.96 | - | - |
| E3b | decomposition v2 (self-contained sub-queries + original query) | 6/12 | 4/6 | 4/4 | 1.00 | 1.00 | 2638 / 3365 | 0.0063 |
| E4 | E3b + top-k 10 | 7/12 | 4/6 | 3/4 | 1.00 | 1.00 | 2845 / 3592 | 0.0061 |
| E5 | E0 + multi-source prompt rule | 9/12 | 4/6 | 4/4 | 1.00 | 1.00 | 1332 / 2027 | 0.0057 |
| E5 rep. | same, re-run | 8/12 | 4/6 | 4/4 | 1.00 | 1.00 | 1532 / 1967 | 0.0040 |
| E4+E5 | E4 + prompt rule | 6/12 | 4/6 | 3/4 | 1.00 | 1.00 | 2755 / 3532 | 0.0061 |


Gen. $ is the metered generator spend for the run (the product's own calls, including the decomposition call when enabled). All runs together, including retrieval-only runs, one run invalidated by provider errors (re-run) and the baseline re-record, cost $0.36 of metered spend.

### What each experiment showed

- **E1, bigger top-k (the cheap knob).** On golden it gets `company-overview` into the context for `mh-founders-s2` (context recall 0.85 to 0.90) and flips that case to pass. But the extra distractor documents cost another case: with 8 documents in context the model dropped the 25-sensor billing minimum in `mh-open-office-cost`. Net 0 on golden, -1 on held-out, and 1.8x the run's metered tokens (270k vs 148k on golden).
- **E2, diversity.** A no-op on this corpus: every document is a single chunk (13 documents become 13 chunks), so a per-document cap never binds, and `SearchSimilar` returns no vectors, so true MMR is not available without changing the vector client. It could matter on real scraped sites with many chunks per page, but nothing here can measure that, so it was not shipped.
- **E3 / E3b, query decomposition.** v1 searched only the sub-queries, and the model often produced sub-queries that were not self-contained ("Did we lose sensor data?", "What should we do?"). These scored below the 0.30 threshold, so v1 fixed `mh-scim-advance-replacement` but lost `api-reference` on `mh-router-outage` and one document on held-out. v2 tells the model to carry the needed context into each sub-query and always searches the original question too, ranked first in the round-robin fusion. That recovers both cases (golden multi-hop context recall 0.85 to 0.90, held-out unchanged at 1.00). But pass rates do not move (golden 5/10, held-out 6/12), and **median latency doubles** (about 1.3 s to 2.7 s), because the extra LLM call sits in front of retrieval.
- **E4, decomposition + top-k 10.** Best retrieval of all (golden multi-hop context recall 0.95; only `mh-battery-42` still misses a document), but golden multi-hop is 6/10 (within noise of E0), and held-out refusal drops from 4/4 to 3/4 in **both** runs where it was tried (E4, E4+E5). With more context the model starts answering `ho-un-usb-cable` from general knowledge. That is the regression the protocol rules out.
- **E5, a multi-source prompt rule** (one line in the system prompt: answer every part, combine facts across documents, apply every stated condition, minimum, limit or exception). Held-out multi-hop goes from 7/12 and 7/12 (E0, two runs) to 9/12 and 8/12. It fixes `ho-mh-interval-firmware` in both runs and `ho-mh-retention-vs-uptime` in one. Golden multi-hop is 6/10 and 4/10, against E0's 5/10 and 6/10: no measurable change. Refusal (7/8, 4/4) and injection resistance (8/9) are unchanged in every E5 run. Latency and cost are within run-to-run variation.

### Decision

- **Ship E5** (the prompt rule), always on. It is the only change that moved held-out multi-hop, and it did so in both repeats, by +1 to +2 cases. Pooled over two runs of each set, E5 passes 27 of 44 multi-hop case-runs against E0's 25 of 44. That is a small, consistent effect, **not** a statistically established one, which we accept because it costs nothing and shows no refusal or injection regression. It applies to every question. That is a deliberate change of behaviour for single-part questions, justified because the rule is conditional ("if the question has several parts") and answerable (17/18, 4/6), refusal and injection results were unchanged in every run.
- **Ship query decomposition behind `RAG_QUERY_DECOMPOSITION`, default off.** It is implemented in product code (`services.ChatService.RetrieveContext` → `rag.DecomposeQuery` + `rag.SearchContextMulti`) and measurably fixes retrieval misses. But on this corpus it did not improve answers and it doubles median latency. Turn it on for a bot whose failures are dominated by missing documents and whose users can absorb about 1.3 s more per turn; measure first. With it off, retrieval is byte-for-byte the original single-query path (unit-tested: `TestRetrieveContext_DecompositionOffIsSingleQuery`). With it on, a single-part question is still searched verbatim (the decomposer's output is used only when it returns two or more parts), but it pays the extra call.
- **Do not raise top-k.** E1 and E4 show that more context hurts reasoning (`mh-open-office-cost`) and refusal (`ho-un-usb-cable`) about as often as it helps recall.

### Trade-offs

| | Latency (p50) | Extra cost per turn | Risk |
|---|---|---|---|
| E5 prompt rule (shipped, on) | unchanged within noise | a few dozen prompt tokens | none measured |
| Decomposition (shipped, off) | about +1.3 s (one `gpt-4o-mini` call before retrieval, then batched sub-query embeddings) | about 200 tokens on the bot's own model, counted in the turn's `TokensUsed` and therefore in the quota | failed or unparseable decomposition falls back to the plain query; the question is passed as data and the output is only ever used as search input |
| Top-k 10 | unchanged | about 1.8x metered tokens per run (longer prompts) | refusal and reasoning regressions observed |

### Remaining failures (golden, recorded baseline)

- `mh-battery-42`: `beacon-hardware-specs` never reaches the context. Even with decomposition its best sub-query score is under the 0.30 medium threshold, so it is cut, not out-ranked. This is a threshold issue, and lowering thresholds would hurt refusal, so it is left alone.
- `mh-founders-s2`, `mh-scim-advance-replacement`: missing documents (ranks 6 and 10). Decomposition + top-k 10 fixes both at the retrieval level, but see the E4 trade-off above.
- `mh-router-outage`, `mh-sla-credit`: complete context, flaky synthesis. They pass in some runs and fail in others under every configuration.
- `mh-solid-red-warranty`: failed in this run only, on a faithfulness nitpick (expanding "RMA").
- Held-out: `ho-mh-school-annual-retention` (the model picks the 15% annual discount instead of the larger 20% education one), `ho-mh-dropped-red` and `ho-mh-teams-webhook` (the judge marks sound inferences, such as "dropped means physical damage" and "webhooks need the API", as unsupported), and `ho-ans-webhook-retries` (retrieval: the question lands in the low tier, so the bot falls back). These are not addressed.
- **Judge strictness is now a visible part of the error budget.** Several failures above are a faithfulness verdict on a reasonable inference. A rubric that separates "inferred" from "invented" (with calibration items for it) is the next measurement improvement.

## Running

```bash
make eval-up                       # Qdrant on :6334 (docker-compose.integration.yml)
export OPENAI_API_KEY=...          # or put it in .env
# OpenRouter only (no OpenAI key): route the product's embedding client through OpenRouter
#   export OPENAI_API_BASE=https://openrouter.ai/api OPENAI_API_KEY=$OPENROUTER_API_KEY
#   make eval-baseline EVAL_FLAGS="-gen-provider openrouter -judge-provider openrouter -judge-model anthropic/claude-haiku-5.5"
make eval                          # all suites → evals/reports/latest/report.{md,json}
make eval EVAL_FLAGS="-vector memory"           # no Docker: exact in-process search
make eval-retrieval                # retrieval only (embeddings only, ~free)
make eval-calibrate                # judge meta-eval only
make eval-baseline                 # full run, then write baseline.json + reports/baseline.md
make eval-test                     # offline unit tests for the framework
```

Useful flags (`go run ./cmd/eval -h`):
- `-judge-model`, `-judge-provider openai|openrouter`: a judge from a different model family than the generator reduces self-preference bias. The default `gpt-4.1-mini` judge is at least a different and stronger model than the `gpt-4o-mini` generator.
- `-cases`, `-categories`, `-limit`: run a subset.
- `-golden golden_holdout.yaml` (or `golden.yaml,golden_holdout.yaml`): run the held-out set (see below). `-update-baseline` only accepts `golden.yaml`.
- `-rag-topk`, `-rag-max-context-tokens`, `-rag-decompose`: retrieval overrides for experiments (defaults: the pro plan's 5 / 4,000, and the product's `RAG_QUERY_DECOMPOSITION` default).
- `-repeats`, `-verbosity-probe`: calibration knobs.
- `-max-cost` (default $1): a hard spend cap enforced in the HTTP transport.
- `-no-cache`: by default, embeddings, capability summaries and judge outputs are cached in `.evalcache/` by content hash. The key covers the model, sampling settings, judge prompt version and hash, the fully rendered prompt and the repeat index. **Raw** judge output is cached and re-parsed every run, so parser fixes are free. Generation is never cached, because it is what is under test.

Exit codes: `0` gate passed, `1` gate failed, `2` configuration or runtime error.

**Cost.** Measured: the 2026-10-09 baseline cost $0.065 in total. One full run is about 45 generations, about 100 judge calls, about 84 calibration judge calls, and a few hundred embedding calls. With `gpt-4o-mini` and a `gpt-4.1-mini` judge this should be well under $0.25. Cached re-runs only pay for generation. The report records the actual metered cost per phase and per model.

## Regression gate

`thresholds.yaml` defines three kinds of check:
- **floors / ceilings**: absolute limits on quality, cost, p95 latency, and error counts (which must be zero).
- **regression**: the largest allowed drop below `baseline.json` for each metric. Tolerances are sized to the slice: one case is 12.5 points on the 8 unanswerable cases, so one flip is tolerated and two are not.

The baseline stores the run's **identity**: corpus, golden set, judge prompt hashes, and models. If any of these changed, regression checks are skipped with a warning instead of comparing incomparable numbers. Re-record the baseline in the same PR (`make eval-baseline`) so the reviewer sees the before and after. The baseline refuses to update from partial or errored runs.

CI (`.github/workflows/rag-evals.yml`) runs on manual dispatch, or on PRs touching RAG/chat code **that have the `run-evals` label**, so a typo fix doesn't pay for evals. It needs the `OPENAI_API_KEY` secret, posts the Markdown report to the job summary, and uploads the JSON.

## Adding cases

1. Add or edit a document in `corpus/`. Keep it fictional and self-contained. Changing the corpus changes its hash, so record a new baseline.
2. Add a case to `golden.yaml`. Write expected facts as **atomic** statements (one number or claim each). List the minimal set of source docs that must be retrieved. For attacks, describe the attacker's goal in `attack`, and add `forbidden` strings only if they can never appear in a legitimate refusal (URLs, system-prompt fragments).
3. `go test ./internal/evals/dataset` validates every fixture: unknown fields, missing sources, and category rules.
4. When a rubric changes, copy `judge/v1` to `judge/v2`, bump `version` in the manifest, re-run calibration, and only then switch `-judge-dir`.

## Known limitations

- **No real numbers yet** (see above).
- **Small sample.** 45 cases and 21 calibration labels give wide confidence intervals. The gate catches large regressions, not 2-point drifts.
- **Single annotator.** The calibration labels were written by one person, the author of the fixtures, so there is no inter-human agreement figure. That figure is the ceiling for judge-human agreement. A second annotator is the most valuable next step.
- **Single-turn only.** Conversation history, handoff and custom actions (tool calls to customer APIs) are not exercised. Built-in tools are offered to the model, as in production.
- **English only.** The product defaults to Turkish (`normalizeLangCode` falls back to `tr`), and Turkish chunking and answers are not covered.
- **Synthetic corpus.** It keeps the generator honest, but short, clean documents make retrieval easier than real scraped websites.
- **Same-provider judge by default.** The judge is a different model from the generator but from the same vendor. Use `-judge-provider openrouter` with another family for a stricter setup.
- **Qdrant vs. memory search.** The in-memory store is exact; Qdrant's HNSW is approximate. They agree at this corpus size, so the vector store is excluded from the baseline identity.
