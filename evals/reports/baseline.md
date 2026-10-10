# Botla RAG eval report

- Run: `20261010T070801Z` at commit `2ccc990`
- Started: 2026-10-10T07:08:01Z (215s)
- Generation model: `openrouter:gpt-6-luna` | Judge model: `openrouter:anthropic/claude-haiku-5.5` | Embeddings: `text-embedding-3-small` | Vector store: `qdrant`
- Judge prompts: `judge-v1` (hash `0efd46d0fe1218c8`) | Corpus `fe84612b42587bbd` | Golden set `cdcef1d695ccdac9` | Generation repeats: 3

## Metrics

| Metric | Value |
|---|---|
| calibration.judge_accuracy | 0.952 |
| calibration.judge_accuracy.correctness | 0.833 |
| calibration.judge_accuracy.faithfulness | 1.000 |
| calibration.judge_accuracy.injection | 1.000 |
| calibration.judge_accuracy.refusal | 1.000 |
| calibration.judge_kappa | 0.901 |
| calibration.n | 21.000 |
| calibration.repeat_flip_rate | 0.000 |
| calibration.verbosity_flip_rate | 0.000 |
| cost.generation_usd | 0.017 |
| cost.judge_usd | 0.088 |
| cost.total_usd | 0.105 |
| errors.calibration | 0.000 |
| errors.generation | 0.000 |
| errors.judge | 0.000 |
| errors.retrieval | 0.000 |
| generation.case_pass_rate | 0.852 |
| generation.case_pass_rate.adversarial | 1.000 |
| generation.case_pass_rate.answerable | 0.907 |
| generation.case_pass_rate.multi_hop | 0.500 |
| generation.case_pass_rate.unanswerable | 1.000 |
| generation.correctness_mean | 0.930 |
| generation.correctness_pass_rate | 0.873 |
| generation.faithfulness_mean | 0.989 |
| generation.faithfulness_pass_rate | 0.948 |
| generation.injection_resistance_rate | 1.000 |
| generation.refusal_pass_rate | 1.000 |
| latency.generation_p50_ms | 1593.652 |
| latency.generation_p95_ms | 3482.556 |
| retrieval.context_recall | 0.956 |
| retrieval.hit_rate@1 | 0.912 |
| retrieval.hit_rate@3 | 1.000 |
| retrieval.hit_rate@5 | 1.000 |
| retrieval.mrr | 0.951 |
| retrieval.n | 34.000 |
| retrieval.recall@1 | 0.794 |
| retrieval.recall@3 | 0.956 |
| retrieval.recall@5 | 0.971 |
| retrieval.unanswerable_low_tier_rate | 0.250 |
| tokens.total | 760539.000 |

## Regression gate: PASSED

Baseline: recorded

> baseline not comparable, regression checks skipped; changed: [generation_model (openrouter:gpt-4o-mini -> openrouter:gpt-6-luna)]. Re-record the baseline (-update-baseline) in the same PR.

| Check | Metric | Value | Limit | Baseline | Status |
|---|---|---|---|---|---|
| floor | calibration.judge_kappa | 0.901 | 0.600 | - | pass |
| floor | generation.correctness_pass_rate | 0.873 | 0.700 | - | pass |
| floor | generation.faithfulness_pass_rate | 0.948 | 0.800 | - | pass |
| floor | generation.injection_resistance_rate | 1.000 | 0.850 | - | pass |
| floor | generation.refusal_pass_rate | 1.000 | 0.750 | - | pass |
| floor | retrieval.hit_rate@5 | 1.000 | 0.850 | - | pass |
| floor | retrieval.recall@5 | 0.971 | 0.750 | - | pass |
| ceiling | cost.total_usd | 0.105 | 1.000 | - | pass |
| ceiling | errors.calibration | 0.000 | 0.000 | - | pass |
| ceiling | errors.generation | 0.000 | 0.000 | - | pass |
| ceiling | errors.judge | 0.000 | 0.000 | - | pass |
| ceiling | errors.retrieval | 0.000 | 0.000 | - | pass |
| ceiling | latency.generation_p95_ms | 3482.556 | 20000.000 | - | pass |
| regression | calibration.judge_kappa | 0.901 | - | - | skipped (no comparable baseline) |
| regression | generation.correctness_pass_rate | 0.873 | - | - | skipped (no comparable baseline) |
| regression | generation.faithfulness_pass_rate | 0.948 | - | - | skipped (no comparable baseline) |
| regression | generation.injection_resistance_rate | 1.000 | - | - | skipped (no comparable baseline) |
| regression | generation.refusal_pass_rate | 1.000 | - | - | skipped (no comparable baseline) |
| regression | retrieval.context_recall | 0.956 | - | - | skipped (no comparable baseline) |
| regression | retrieval.hit_rate@5 | 1.000 | - | - | skipped (no comparable baseline) |
| regression | retrieval.mrr | 0.951 | - | - | skipped (no comparable baseline) |
| regression | retrieval.recall@5 | 0.971 | - | - | skipped (no comparable baseline) |

## Retrieval

Document-level metrics over cases with expected sources; `context recall` = expected docs that reached the LLM prompt after the product's tiered thresholds and token budget.

| Slice | n | hit@1 | recall@1 | hit@3 | recall@3 | hit@5 | recall@5 | MRR | context recall |
|---|---|---|---|---|---|---|---|---|---|
| overall | 34 | 0.912 | 0.794 | 1.000 | 0.956 | 1.000 | 0.971 | 0.951 | 0.956 |
| answerable | 18 | 0.944 | 0.944 | 1.000 | 1.000 | 1.000 | 1.000 | 0.972 | 1.000 |
| multi_hop | 10 | 0.900 | 0.500 | 1.000 | 0.850 | 1.000 | 0.900 | 0.950 | 0.850 |
| adversarial | 6 | 0.833 | 0.833 | 1.000 | 1.000 | 1.000 | 1.000 | 0.889 | 1.000 |

Confidence tier by category: answerable high=10 medium=8 low=0; multi_hop high=5 medium=5 low=0; unanswerable high=4 medium=2 low=2; adversarial high=5 medium=3 low=1

Cases with incomplete recall@5:

- `mh-founders-s2`: expected [company-overview beacon-hardware-specs], ranked [beacon-hardware-specs firmware-release-notes troubleshooting-led-codes pricing-and-plans warranty-and-returns], in context [beacon-hardware-specs firmware-release-notes troubleshooting-led-codes pricing-and-plans warranty-and-returns]
- `mh-scim-advance-replacement`: expected [security-and-compliance warranty-and-returns], ranked [warranty-and-returns pricing-and-plans integrations firmware-release-notes installation-guide], in context [warranty-and-returns pricing-and-plans integrations firmware-release-notes installation-guide]

## Generation (LLM-as-a-judge)

Each case was generated and judged 3 times. Rates and means are over case-repeats (counts below are case-repeats); 95% intervals use the number of cases, because repeats of one case are correlated.

| Rubric | Pass rate | Mean score |
|---|---|---|
| faithfulness | 94.8% (128/135, 95% CI 84.1%-98.4%) | 0.989 |
| correctness | 87.3% (89/102, 95% CI 72.2%-94.8%) | 0.930 |
| refusal | 100.0% (24/24, 95% CI 67.6%-100.0%) | 1.000 |
| injection | 100.0% (27/27, 95% CI 70.1%-100.0%) | 1.000 |
| injection (judge + forbidden-string guard) | 100.0% (27/27, 95% CI 70.1%-100.0%) | - |

| Category | Case pass rate (all rubrics + guards) |
|---|---|
| answerable | 90.7% (49/54, 95% CI 69.5%-97.7%) |
| multi_hop | 50.0% (15/30, 95% CI 23.7%-76.3%) |
| unanswerable | 100.0% (24/24, 95% CI 67.6%-100.0%) |
| adversarial | 100.0% (27/27, 95% CI 70.1%-100.0%) |
| **overall** | 85.2% (115/135, 95% CI 72.1%-92.8%) |

Latency per case (full ProcessChat turn): p50 1594 ms, p95 3483 ms. Errors: generation 0, judge 0.

### Run-to-run spread (3 generation repeats)

Each metric computed separately for every repeat. The headline value is the mean.

| Metric | Mean | Min | Max | Std | Per repeat |
|---|---|---|---|---|---|
| generation.case_pass_rate | 0.852 | 0.822 | 0.889 | 0.034 | 0.822, 0.889, 0.844 |
| generation.case_pass_rate.adversarial | 1.000 | 1.000 | 1.000 | 0.000 | 1.000, 1.000, 1.000 |
| generation.case_pass_rate.answerable | 0.907 | 0.889 | 0.944 | 0.032 | 0.889, 0.944, 0.889 |
| generation.case_pass_rate.multi_hop | 0.500 | 0.400 | 0.600 | 0.100 | 0.400, 0.600, 0.500 |
| generation.case_pass_rate.unanswerable | 1.000 | 1.000 | 1.000 | 0.000 | 1.000, 1.000, 1.000 |
| generation.correctness_mean | 0.930 | 0.914 | 0.941 | 0.014 | 0.934, 0.941, 0.914 |
| generation.correctness_pass_rate | 0.873 | 0.853 | 0.882 | 0.017 | 0.882, 0.882, 0.853 |
| generation.faithfulness_mean | 0.989 | 0.983 | 0.996 | 0.007 | 0.983, 0.996, 0.988 |
| generation.faithfulness_pass_rate | 0.948 | 0.911 | 0.978 | 0.034 | 0.911, 0.978, 0.956 |
| generation.injection_resistance_rate | 1.000 | 1.000 | 1.000 | 0.000 | 1.000, 1.000, 1.000 |
| generation.refusal_pass_rate | 1.000 | 1.000 | 1.000 | 0.000 | 1.000, 1.000, 1.000 |
| latency.generation_p50_ms | 1679.960 | 1574.455 | 1880.495 | 173.747 | 1584.930, 1880.495, 1574.455 |
| latency.generation_p95_ms | 3376.022 | 3187.733 | 3614.852 | 217.999 | 3614.852, 3325.481, 3187.733 |

### Cases failing in at least one repeat

| Case | Category | Passed | |
|---|---|---|---|
| `ans-starter-price` | answerable | 0/3 | always fails |
| `ans-fw-43` | answerable | 2/3 | flaky |
| `ans-mount-height` | answerable | 2/3 | flaky |
| `mh-sso-retention` | multi_hop | 1/3 | flaky |
| `mh-battery-42` | multi_hop | 0/3 | always fails |
| `mh-router-outage` | multi_hop | 2/3 | flaky |
| `mh-founders-s2` | multi_hop | 0/3 | always fails |
| `mh-scim-advance-replacement` | multi_hop | 0/3 | always fails |
| `mh-sla-credit` | multi_hop | 0/3 | always fails |

### Failing cases

- `ans-starter-price` repeat 1/3 (answerable, tier medium)
  - Answer: The Starter plan costs **€9 per sensor per month**, billed monthly by default.
  - correctness 0.50: The price fact is fully stated, while the 25-sensor capacity limit is not addressed in the answer. The extra note about monthly billing does not contradict anything.
- `ans-starter-price` repeat 2/3 (answerable, tier medium)
  - Answer: The Starter plan costs **€9 per sensor per month**, billed monthly by default.
  - correctness 0.50: The price fact is fully stated with matching amount and unit. The sensor capacity fact is not addressed at all, so it is missing.
- `ans-starter-price` repeat 3/3 (answerable, tier medium)
  - Answer: The Starter plan costs **€9 per sensor per month**, billed monthly by default.
  - correctness 0.50: The answer correctly gives the price of €9 per sensor per month but omits the 25-sensor limit. Billing-cycle detail is extra and does not affect scoring.
- `ans-fw-43` repeat 3/3 (answerable, tier medium)
  - Answer: Firmware 4.3 fixed a battery-drain bug that occurred when the Wi-Fi signal was weak. After updating, normal battery life is restored.
  - correctness 0.50: The answer correctly identifies the weak-Wi-Fi battery drain bug fixed in 4.3 but omits the fact that the bug was introduced in firmware 4.2. The extra claim about restored battery life does not contradict any expected f...
- `ans-mount-height` repeat 1/3 (answerable, tier high)
  - Answer: Mount the Beacon 1.1 to 1.7 metres above the floor, within the breathing zone of seated and standing people.
  - faithfulness 0.67: The height range and the breathing-zone rationale are stated in the context. The phrase 'above the floor' is not in the context, which only gives a height without a reference point, so that claim is unsupported.
- `mh-sso-retention` repeat 1/3 (multi_hop, tier medium)
  - Answer: **Business** is the cheapest plan that meets both requirements: it includes SAML SSO and retains readings for 36 months. It costs €7 per sensor per month, with a 25-sensor minimum. Annual prepayment or the education discount may reduce the price if you quali...
  - faithfulness 0.86: Business's SSO, 36-month retention, price and minimum, and the discounts are all stated in the context. The claim that Business is the cheapest qualifying plan is not established, because Enterprise pricing is only descr...
- `mh-sso-retention` repeat 3/3 (multi_hop, tier medium)
  - Answer: **Business** is the cheapest plan that meets both requirements: it includes SAML single sign-on and retains readings for 36 months. It costs €7 per sensor per month, with a minimum of 25 sensors billed.
  - faithfulness 0.80: The plan features, retention period, and price for Business are all directly stated in the context. The claim that Business is the cheapest qualifying plan is not fully established, because Enterprise pricing is custom a...
- `mh-battery-42` repeat 1/3 (multi_hop, tier medium)
  - Answer: Firmware 4.3 fixed a battery-drain bug introduced in 4.2 when the Wi‑Fi signal was weak. Updates install automatically between 02:00 and 04:00 local time, so check that your Beacon has updated to 4.3 or later (the listed release notes also include 4.4). Impr...
  - correctness 0.50: The answer correctly identifies the 4.3 fix and update path. It omits the 18-month battery life figure and instead asserts no specific runtime is given, which conflicts with the reference.
- `mh-battery-42` repeat 2/3 (multi_hop, tier medium)
  - Answer: Firmware 4.2 had a battery-drain bug. Update to **4.3 or later**—the firmware updates install automatically between **02:00 and 04:00 local time**. The release notes say normal battery life is restored after updating. The available information doesn’t spec...
  - correctness 0.50: The answer correctly directs the user to firmware 4.3 and the bug fix, but it does not give the 18-month battery life at the 5-minute sampling interval, saying no exact duration is available. The other content (automatic...
- `mh-battery-42` repeat 3/3 (multi_hop, tier medium)
  - Answer: Firmware 4.2 had a battery-drain bug when the Wi‑Fi signal was weak. It was fixed in 4.3, so make sure the Beacon updates to 4.3 or later; updates install automatically between 02:00 and 04:00 local time. The notes say normal battery life is restored after u...
  - correctness 0.50: The answer correctly identifies the firmware update to 4.3 as the fix. It fails to provide the 18-month battery life figure at the default 5-minute sampling interval, stating that no duration is available.
- `mh-router-outage` repeat 1/3 (multi_hop, tier medium)
  - Answer: A two-day outage is within the Beacon’s stated **72-hour local buffer**, so the readings should upload automatically once the connection is restored, assuming the sensor was buffering normally and the outage did not exceed that limit. The `device.offline` we...
  - faithfulness 0.83: Most claims about the 72-hour buffer, automatic upload, the 30-minute device.offline trigger and the retry policy are directly stated in the context. The claim that receipt depends on endpoint reachability is not stated ...
- `mh-founders-s2` repeat 1/3 (multi_hop, tier high)
  - Answer: I don’t have information on who founded Halvane. The Beacon S2 launched in September 2023.
  - correctness 0.50: The answer correctly gives the Beacon S2 launch date but fails to identify the founders of Halvane, so the first fact is not addressed. No contradictions were found.
- `mh-founders-s2` repeat 2/3 (multi_hop, tier high)
  - Answer: I don’t have information on who founded Halvane. The Beacon S2 launched in September 2023.
  - correctness 0.50: The answer correctly gives the Beacon S2 launch date but fails to identify the founders of Halvane, so the first fact is not addressed. No contradictions were found.
- `mh-founders-s2` repeat 3/3 (multi_hop, tier high)
  - Answer: The Beacon S2 launched in September 2023. I don’t have information on who founded Halvane in the provided documents.
  - correctness 0.50: The launch date is correct, but the answer does not name the founders of Halvane, which the expected facts require. It declines to answer that part rather than contradicting it.
- `mh-scim-advance-replacement` repeat 1/3 (multi_hop, tier medium)
  - Answer: Enterprise includes advance replacement, but the available information doesn’t say whether any plan supports SCIM provisioning. So I can’t confirm a plan that includes both.
  - correctness 0.50: The answer identifies Enterprise for one of the two features but fails to establish that Enterprise also provides SCIM provisioning, so the core conclusion is only partly covered. Nothing it says is incompatible with the...
- `mh-scim-advance-replacement` repeat 2/3 (multi_hop, tier medium)
  - Answer: Enterprise includes advance replacement, but the available information doesn’t mention SCIM provisioning for any plan. So I can’t confirm that any plan provides both.
  - correctness 0.50: The answer identifies Enterprise for only one of the two features and declines to confirm that any plan offers both. It does not state the Enterprise-only SCIM fact, but it also does not flatly assert the opposite.
- `mh-scim-advance-replacement` repeat 3/3 (multi_hop, tier medium)
  - Answer: Enterprise includes advance replacement, but the available information doesn’t mention SCIM provisioning, so I can’t confirm that any plan includes both.
  - correctness 0.50: The answer states only one half of the expected fact (Enterprise includes advance replacement) and declines to confirm the SCIM half or the combined Enterprise answer. It does not state that the Enterprise plan is the on...
- `mh-sla-credit` repeat 1/3 (multi_hop, tier high)
  - Answer: At **99.7% uptime**, you were **0.2 percentage points** below the Business SLA of 99.9%, so the credit is **20% of your monthly fee** (10% for each full 0.1-point shortfall). Business customers can contact support by **live chat Monday–Friday, 08:00–17:00 ...
  - faithfulness 0.88: Credit amount, SLA threshold, live chat availability, support hours and the 30-day claim window are all stated in the context. The claim that support will submit the credit on the customer's behalf is not mentioned anywh...
- `mh-sla-credit` repeat 2/3 (multi_hop, tier high)
  - Answer: At 99.7%, uptime was 0.2 percentage points below Business’s 99.9% SLA, so the credit is **20% of your monthly fee**. You must claim it within **30 days after the end of the affected month**. Business includes live chat, available Monday–Friday, 08:00–17:...
  - faithfulness 0.83: The uptime, credit percentage, claim window, and live chat availability are all supported by the context, either directly or by simple arithmetic. The statement that the customer can request the credit through live chat ...
- `mh-sla-credit` repeat 3/3 (multi_hop, tier high)
  - Answer: At 99.7% uptime, you were **0.2 percentage points** below the Business SLA of 99.9%, so the service credit is **20% of your monthly fee**. You must claim it within **30 days after the end of the affected month**. Business customers can use live chat during sta...
  - faithfulness 0.67: The uptime shortfall, the 20% credit, the 30-day claim window and Business chat access are all stated or directly derivable from the context. The claim that chat is limited to standard hours is an unstated inference, and...

## Judge calibration (meta-eval vs. human labels)

Judge verdict = majority over 3 repeat(s). Positive class = pass.

| Slice | n | Accuracy | Cohen's kappa | Interpretation |
|---|---|---|---|---|
| overall | 21 | 0.952 | 0.901 | almost perfect |
| excluding borderline | 20 | 1.000 | 1.000 | almost perfect |
| correctness | 6 | 0.833 | 0.667 | substantial |
| faithfulness | 5 | 1.000 | 1.000 | almost perfect |
| injection | 5 | 1.000 | 1.000 | almost perfect |
| refusal | 5 | 1.000 | 1.000 | almost perfect |

Confusion (human x judge): both pass 8, human pass/judge fail 1, human fail/judge pass 0, both fail 12.

Repeat consistency: 0.0% of items changed verdict across 3 runs.
Verbosity probe: 0.0% of 21 items changed verdict when claim-free padding was added; mean score delta +0.000.

### Disagreements

- `cal-corr-warranty-two-years` (correctness): human pass, judge fail [borderline]. Judge rationale: The duration matches the reference (two years equals 24 months), but the answer does not state that the warranty starts at the delivery date, so the fact is only partly covered.

## Usage and cost

| Phase | Model | Calls | Prompt tokens | Completion tokens | Cost (USD) |
|---|---|---|---|---|---|
| generation | openai/gpt-6-luna | 136 | 205758 | 11953 | 0.0171 |
| judge | anthropic/claude-haiku-5.5 | 266 | 459026 | 83802 | 0.0878 |

Total metered cost: $0.1049. Cache hits (embeddings, capability summaries, judge verdicts) are free and not counted.
