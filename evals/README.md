# Botla RAG evaluation suite

An offline-reproducible evaluation harness for Botla's RAG chatbot. It measures retrieval quality, uses an LLM judge to check answer faithfulness (hallucination), correctness, refusal and prompt-injection resistance, measures how far that judge can be trusted, and fails CI when quality regresses.

> **Status: no baseline recorded yet.** The framework, fixtures and tests are complete. The one real run attempted while building this failed preflight because the OpenAI and OpenRouter keys available at the time were rejected (HTTP 401). No metric in this repository came from a real model. `evals/baseline.json` says `not_recorded` until someone runs `make eval-baseline` with valid keys. See [Results](#results).

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

**Not yet recorded.** The labels were written before any judge run, so there is nothing to report and no number has been filled in by hand. After the first real run, `evals/reports/baseline.md` holds the full report (per-case answers, judge rationales, disagreements, cost). Summarise kappa here and link it.

A smoke run with an offline fake provider confirmed that the whole pipeline executes: ingestion into real Qdrant, retrieval, `ProcessChat`, judge, calibration, report and gate. Its numbers are meaningless by construction and were discarded.

## Running

```bash
make eval-up                       # Qdrant on :6334 (docker-compose.integration.yml)
export OPENAI_API_KEY=...          # or put it in .env
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
- `-repeats`, `-verbosity-probe`: calibration knobs.
- `-max-cost` (default $1): a hard spend cap enforced in the HTTP transport.
- `-no-cache`: by default, embeddings, capability summaries and judge outputs are cached in `.evalcache/` by content hash. The key covers the model, sampling settings, judge prompt version and hash, the fully rendered prompt and the repeat index. **Raw** judge output is cached and re-parsed every run, so parser fixes are free. Generation is never cached, because it is what is under test.

Exit codes: `0` gate passed, `1` gate failed, `2` configuration or runtime error.

**Cost.** Estimated, not measured: one full run is about 45 generations, about 100 judge calls, about 84 calibration judge calls, and a few hundred embedding calls. With `gpt-4o-mini` and a `gpt-4.1-mini` judge this should be well under $0.25. Cached re-runs only pay for generation. The report records the actual metered cost per phase and per model.

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
