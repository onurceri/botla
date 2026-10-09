// Command eval runs Botla's RAG evaluation suite: retrieval metrics,
// LLM-as-a-judge generation scoring and judge calibration, followed by the
// regression gate. See evals/README.md.
//
// Exit codes: 0 = gate passed, 1 = gate failed, 2 = configuration/runtime error.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/onurceri/botla-app/internal/evals/cache"
	"github.com/onurceri/botla-app/internal/evals/dataset"
	"github.com/onurceri/botla-app/internal/evals/gate"
	"github.com/onurceri/botla-app/internal/evals/harness"
	"github.com/onurceri/botla-app/internal/evals/judge"
	"github.com/onurceri/botla-app/internal/evals/runner"
	"github.com/onurceri/botla-app/internal/evals/usage"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/rag"
	"github.com/onurceri/botla-app/pkg/config"
	"github.com/onurceri/botla-app/pkg/langconfig"
	"github.com/onurceri/botla-app/pkg/logger"
)

const (
	embeddingModel = "text-embedding-3-small" // hard-coded in rag.OpenAIClient
	evalChatbotID  = "eval-halvane-bot"
	evalUserID     = "eval-user"
)

type options struct {
	suites, evalsDir, judgeDir, outDir, cacheDir       string
	vector, qdrantURL, qdrantKey, collection           string
	genProvider, genModel, judgeProvider, judgeModel   string
	genTemperature, judgeTemperature, maxCost          float64
	concurrency, repeats, limit                        int
	cases, categories, logLevel, golden                string
	verbosityProbe, noCache, updateBaseline, preflight bool
	// Retrieval overrides (default: the seeded pro plan and the product's env defaults).
	ragTopK, ragMaxContextTokens int
	ragDecompose                 bool
}

func parseFlags() options {
	var o options
	flag.StringVar(&o.suites, "suites", strings.Join(runner.AllSuites, ","), "comma-separated suites: retrieval,generation,calibration")
	flag.StringVar(&o.evalsDir, "evals-dir", "evals", "directory with corpus/, golden.yaml, calibration/, thresholds.yaml, baseline.json, pricing.yaml")
	flag.StringVar(&o.golden, "golden", "golden.yaml", "comma-separated golden set files in <evals-dir> (e.g. golden_holdout.yaml, or golden.yaml,golden_holdout.yaml)")
	flag.StringVar(&o.judgeDir, "judge-dir", "", "judge prompt set directory (default <evals-dir>/judge/v1)")
	flag.StringVar(&o.outDir, "out", "", "report output directory (default <evals-dir>/reports/latest)")
	flag.StringVar(&o.cacheDir, "cache-dir", ".evalcache", "content-hash cache for embeddings, capability summaries and judge verdicts")
	flag.BoolVar(&o.noCache, "no-cache", false, "disable the cache (every call is paid again)")
	flag.StringVar(&o.vector, "vector", "qdrant", "vector store: qdrant (real) or memory (exact in-process search)")
	flag.StringVar(&o.qdrantURL, "qdrant-url", envOr("EVAL_QDRANT_URL", "http://localhost:6334"), "Qdrant URL (default: integration stack from docker-compose.integration.yml)")
	flag.StringVar(&o.qdrantKey, "qdrant-api-key", os.Getenv("EVAL_QDRANT_API_KEY"), "Qdrant API key, if any")
	flag.StringVar(&o.collection, "collection", "botla_eval", "Qdrant collection used for the eval (never the product collection)")
	flag.StringVar(&o.genProvider, "gen-provider", "openai", "generation provider: openai or openrouter")
	flag.StringVar(&o.genModel, "gen-model", config.DefaultModelName, "chatbot model (the product's bot.Model value)")
	flag.Float64Var(&o.genTemperature, "gen-temperature", 0, "chatbot temperature (product default is 0.7; 0 reduces run-to-run variance)")
	flag.StringVar(&o.judgeProvider, "judge-provider", "openai", "judge provider: openai or openrouter")
	flag.StringVar(&o.judgeModel, "judge-model", "gpt-4.1-mini", "judge model")
	flag.Float64Var(&o.judgeTemperature, "judge-temperature", 0, "judge sampling temperature")
	flag.IntVar(&o.concurrency, "concurrency", 4, "parallel cases")
	flag.IntVar(&o.repeats, "repeats", 3, "judge repeats per calibration item (self-consistency check)")
	flag.BoolVar(&o.verbosityProbe, "verbosity-probe", true, "re-judge calibration items with claim-free padding to detect verbosity bias")
	flag.Float64Var(&o.maxCost, "max-cost", 1.0, "abort new provider calls once metered spend reaches this many USD (0 = unlimited)")
	flag.StringVar(&o.cases, "cases", "", "only run these case IDs (comma-separated)")
	flag.StringVar(&o.categories, "categories", "", "only run these categories (comma-separated)")
	flag.IntVar(&o.limit, "limit", 0, "run at most N cases")
	flag.BoolVar(&o.updateBaseline, "update-baseline", false, "write this run's metrics to baseline.json (requires a full, error-free run)")
	flag.BoolVar(&o.preflight, "preflight", true, "make one cheap call per provider before starting, to fail fast on bad keys")
	flag.StringVar(&o.logLevel, "log-level", "ERROR", "product logger level (logs go to stdout)")
	strategy := rag.RetrievalStrategyFromEnv()
	flag.IntVar(&o.ragTopK, "rag-topk", 5, "chunks admitted to the context (the pro plan's chat_rag_top_k)")
	flag.IntVar(&o.ragMaxContextTokens, "rag-max-context-tokens", 4000, "context token budget (the pro plan's chat_rag_max_context_tokens)")
	flag.BoolVar(&o.ragDecompose, "rag-decompose", strategy.Decompose, "query decomposition for multi-part questions (default: product default / "+rag.EnvQueryDecomposition+")")
	flag.Parse()
	if o.judgeDir == "" {
		o.judgeDir = filepath.Join(o.evalsDir, "judge", "v1")
	}
	if o.outDir == "" {
		o.outDir = filepath.Join(o.evalsDir, "reports", "latest")
	}
	return o
}

func main() {
	_ = godotenv.Load() // best-effort: keys may also come from the environment (CI secrets)
	o := parseFlags()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	code, err := run(ctx, o)
	stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, "eval:", err)
	}
	os.Exit(code)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func has(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

type fixtures struct {
	corpus      *dataset.Corpus
	golden      *dataset.GoldenSet
	calibration *dataset.CalibrationSet
	thresholds  *gate.Thresholds
	baseline    *gate.Baseline
	pricing     *usage.Pricing
	prompts     *judge.PromptSet
}

func loadFixtures(o options) (*fixtures, error) {
	var f fixtures
	var err error
	if f.corpus, err = dataset.LoadCorpus(filepath.Join(o.evalsDir, "corpus")); err != nil {
		return nil, err
	}
	var goldenPaths []string
	for _, name := range splitList(o.golden) {
		goldenPaths = append(goldenPaths, filepath.Join(o.evalsDir, name))
	}
	if f.golden, err = dataset.LoadGoldenFiles(goldenPaths...); err != nil {
		return nil, err
	}
	if err = f.golden.Validate(f.corpus); err != nil {
		return nil, fmt.Errorf("invalid golden set: %w", err)
	}
	if f.calibration, err = dataset.LoadCalibration(filepath.Join(o.evalsDir, "calibration", "human_labels.yaml")); err != nil {
		return nil, err
	}
	if err = f.calibration.Validate(f.corpus); err != nil {
		return nil, fmt.Errorf("invalid calibration set: %w", err)
	}
	if f.thresholds, err = gate.LoadThresholds(filepath.Join(o.evalsDir, "thresholds.yaml")); err != nil {
		return nil, err
	}
	if f.baseline, err = gate.LoadBaseline(filepath.Join(o.evalsDir, "baseline.json")); err != nil {
		return nil, err
	}
	if f.pricing, err = usage.LoadPricing(filepath.Join(o.evalsDir, "pricing.yaml")); err != nil {
		return nil, err
	}
	if f.prompts, err = judge.LoadPromptSet(o.judgeDir); err != nil {
		return nil, err
	}
	return &f, nil
}

func providerBase(provider string) (base, key string, err error) {
	switch provider {
	case "openai":
		base = strings.TrimRight(envOr("OPENAI_API_BASE", "https://api.openai.com"), "/")
		if !strings.HasSuffix(base, "/v1") {
			base += "/v1"
		}
		return base, os.Getenv("OPENAI_API_KEY"), nil
	case "openrouter":
		return envOr("OPENROUTER_API_BASE", "https://openrouter.ai/api/v1"), os.Getenv("OPENROUTER_API_KEY"), nil
	default:
		return "", "", fmt.Errorf("unknown provider %q (use openai or openrouter)", provider)
	}
}

//nolint:gocyclo // linear wiring of the eval run; splitting it would obscure the sequence
func run(ctx context.Context, o options) (int, error) {
	suites := splitList(o.suites)
	for _, s := range suites {
		if !has(runner.AllSuites, s) {
			return 2, fmt.Errorf("unknown suite %q", s)
		}
	}
	fx, err := loadFixtures(o)
	if err != nil {
		return 2, err
	}
	var cats []dataset.Category
	for _, c := range splitList(o.categories) {
		cats = append(cats, dataset.Category(c))
	}
	cases := fx.golden.Filter(splitList(o.cases), cats, o.limit)
	needPipeline := has(suites, runner.SuiteRetrieval) || has(suites, runner.SuiteGeneration)
	needJudge := has(suites, runner.SuiteGeneration) || has(suites, runner.SuiteCalibration)

	log := logger.New(o.logLevel)
	budget := usage.NewBudget(o.maxCost)
	metered := &http.Client{Timeout: 90 * time.Second, Transport: &usage.Transport{Base: http.DefaultTransport, Pricing: fx.pricing, Budget: budget}}
	cacheDir := o.cacheDir
	if o.noCache {
		cacheDir = ""
	}
	store := cache.New(cacheDir)

	identity := gate.Identity{
		CorpusHash: fx.corpus.Hash(), GoldenHash: fx.golden.Hash(),
		JudgeVersion: fx.prompts.Version, JudgeHash: fx.prompts.Hash,
		EmbedModel: embeddingModel, VectorStore: o.vector,
	}
	deps := runner.Deps{}

	if needJudge {
		base, key, perr := providerBase(o.judgeProvider)
		if perr != nil {
			return 2, perr
		}
		if key == "" {
			return 2, fmt.Errorf("judge provider %s: API key is not set", o.judgeProvider)
		}
		jc := &judge.HTTPClient{BaseURL: base, APIKey: key, Model: o.judgeModel, Temperature: o.judgeTemperature, MaxTokens: 1500, JSONMode: true, HTTP: metered}
		if o.preflight {
			pctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			_, perr = jc.Complete(pctx, "Reply with the JSON object {\"ok\": true}.", "ping")
			cancel()
			if perr != nil {
				return 2, fmt.Errorf("judge preflight failed (check %s key/model): %w", o.judgeProvider, perr)
			}
		}
		deps.Grader = &judge.Judge{Prompts: fx.prompts, Client: jc, Cache: store}
		identity.JudgeModel = o.judgeProvider + ":" + o.judgeModel
	}

	if needPipeline {
		pipe, perr := buildPipeline(ctx, o, fx, metered, store, log)
		if perr != nil {
			return 2, perr
		}
		deps.Retriever, deps.Generator = pipe, pipe
		identity.GenModel = o.genProvider + ":" + o.genModel
	}

	en := langconfig.Get("en").UserMessages
	cfg := runner.Config{
		Suites: suites, Cases: cases, Corpus: fx.corpus, Calibration: fx.calibration,
		Concurrency: o.concurrency, JudgeRepeats: o.repeats, VerbosityProbe: o.verbosityProbe,
		ErrorMessages: []string{en.ErrorMessage, en.EmptyStateMessage},
		Identity:      identity, Thresholds: fx.thresholds, Baseline: fx.baseline,
	}
	rep, err := runner.Run(ctx, cfg, deps)
	if err != nil {
		return 2, err
	}
	rep.RunID = rep.StartedAt.Format("20060102T150405Z")
	rep.Commit = gitCommit()
	jsonPath, mdPath, err := rep.WriteFiles(o.outDir)
	if err != nil {
		return 2, err
	}
	fmt.Println(rep.Summary())
	fmt.Printf("Reports: %s, %s\nMetered spend, all calls incl. ingestion and preflight: $%.4f\n", jsonPath, mdPath, budget.Spent())
	if budget.Exceeded() {
		return 2, fmt.Errorf("cost budget of $%.2f was reached; results are incomplete", o.maxCost)
	}

	if o.updateBaseline {
		if err := writeBaseline(o, rep, len(cases) == len(fx.golden.Cases), suites); err != nil {
			return 2, err
		}
		fmt.Println("Baseline updated:", filepath.Join(o.evalsDir, "baseline.json"))
	}
	if !rep.Gate.Passed {
		return 1, errors.New("regression gate failed")
	}
	return 0, nil
}

func buildPipeline(ctx context.Context, o options, fx *fixtures, metered *http.Client, store *cache.Cache, log *logger.Logger) (*harness.Pipeline, error) {
	pcfg := &config.Config{
		OPENAI_API_KEY:  os.Getenv("OPENAI_API_KEY"),
		OPENAI_API_BASE: os.Getenv("OPENAI_API_BASE"),
	}
	if pcfg.OPENAI_API_KEY == "" {
		return nil, errors.New("OPENAI_API_KEY is required (the product embeds with OpenAI)")
	}
	oai, err := rag.NewOpenAIClient(pcfg, rag.WithHTTPClient(metered))
	if err != nil {
		return nil, fmt.Errorf("openai client: %w", err)
	}
	if o.genProvider == "openrouter" {
		pcfg.OPENROUTER_API_KEY = os.Getenv("OPENROUTER_API_KEY")
		pcfg.OPENROUTER_API_BASE = os.Getenv("OPENROUTER_API_BASE")
		if pcfg.OPENROUTER_API_KEY == "" {
			return nil, errors.New("OPENROUTER_API_KEY is required for -gen-provider openrouter")
		}
	}
	factory := rag.NewClientFactory(pcfg)
	factory.RegisterClient("openai", oai)
	if o.genProvider == "openrouter" {
		orc, orErr := rag.NewOpenRouterClient(pcfg, rag.WithOpenRouterHTTPClient(metered))
		if orErr != nil {
			return nil, fmt.Errorf("openrouter client: %w", orErr)
		}
		factory.RegisterClient("openrouter", orc)
	}
	embedder := &harness.CachingEmbedder{Inner: oai, Cache: store, Model: embeddingModel}
	if o.preflight {
		pctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		_, err = oai.CreateEmbedding(pctx, "preflight")
		cancel()
		if err != nil {
			return nil, fmt.Errorf("embedding preflight failed (check OPENAI_API_KEY): %w", err)
		}
	}

	var vector rag.VectorClient
	switch o.vector {
	case "qdrant":
		qc, qerr := rag.NewQdrantClient(&rag.QdrantConfig{URL: o.qdrantURL, APIKey: o.qdrantKey, CollectionName: o.collection})
		if qerr != nil {
			return nil, fmt.Errorf("qdrant client: %w", qerr)
		}
		vector = qc
	case "memory":
		vector = harness.NewMemoryVectorStore()
	default:
		return nil, fmt.Errorf("unknown vector store %q", o.vector)
	}

	ing := &harness.Ingestor{Embedder: embedder, Vector: vector, LLM: oai, Cache: store, Log: log, Language: "en"}
	ictx := usage.WithMeter(ctx, usage.NewMeter())
	index, err := ing.Ingest(ictx, evalChatbotID, fx.corpus)
	if err != nil {
		return nil, fmt.Errorf("ingest corpus (is Qdrant running? try `make eval-up` or -vector memory): %w", err)
	}
	fmt.Printf("Ingested %d documents into %d chunks (%s)\n", len(fx.corpus.Docs), index.TotalChunks(), o.vector)
	fmt.Printf("Retrieval: top-k %d, context budget %d tokens, query decomposition %t\n", o.ragTopK, o.ragMaxContextTokens, o.ragDecompose)

	return &harness.Pipeline{
		Factory:  factory,
		Embedder: embedder,
		Vector:   vector,
		Index:    index,
		Log:      log,
		Bot: models.Chatbot{
			ID: evalChatbotID, UserID: evalUserID, Name: "Halvane Assistant",
			LanguageCode: "en", Model: o.genModel,
			Temperature: float32(o.genTemperature), MaxTokens: 512, // product default max_tokens
		},
		// Mirrors the seeded "pro" plan (db/migrations/000001_init.up.sql).
		Plan: models.Plan{Code: "pro", Limits: &models.PlanLimits{
			ChatRAGTopK: o.ragTopK, ChatRAGMaxContextTokens: o.ragMaxContextTokens,
			GuardrailsCanUseSmartFallback: true, GuardrailsCanCustomizeThresholds: true,
		}},
		Retrieval: rag.RetrievalStrategy{Decompose: o.ragDecompose},
	}, nil
}

func writeBaseline(o options, rep *runner.Report, fullSet bool, suites []string) error {
	if o.golden != "golden.yaml" {
		return errors.New("refusing to update the baseline from a run that is not over golden.yaml alone (held-out sets are for checking, not gating)")
	}
	if !fullSet || len(suites) != len(runner.AllSuites) {
		return errors.New("refusing to update the baseline from a partial run (filters or suite subset)")
	}
	for _, k := range []string{"errors.retrieval", "errors.generation", "errors.judge", "errors.calibration"} {
		if rep.Metrics[k] > 0 {
			return fmt.Errorf("refusing to update the baseline: %s = %.0f", k, rep.Metrics[k])
		}
	}
	b := &gate.Baseline{
		Status:     gate.BaselineRecorded,
		RecordedAt: rep.FinishedAt.Format(time.RFC3339),
		Commit:     rep.Commit,
		Identity:   rep.Identity,
		Metrics:    rep.Metrics,
		Note:       "Recorded with `go run ./cmd/eval -update-baseline`. Full report: evals/reports/baseline.md",
	}
	if err := b.Write(filepath.Join(o.evalsDir, "baseline.json")); err != nil {
		return err
	}
	md, err := os.ReadFile(filepath.Join(o.outDir, "report.md")) // #nosec G304 -- report we just wrote
	if err != nil {
		return fmt.Errorf("read report for baseline copy: %w", err)
	}
	if err := os.WriteFile(filepath.Join(o.evalsDir, "reports", "baseline.md"), md, 0o600); err != nil {
		return fmt.Errorf("write baseline report: %w", err)
	}
	return nil
}

func gitCommit() string {
	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
