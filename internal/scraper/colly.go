package scraper

import (
    "math/rand"
    "net/http"
    "time"
    "os"
    "strings"

    "github.com/gocolly/colly"
    "github.com/gocolly/colly/queue"
    "github.com/onurceri/botla-co/pkg/logger"
)

type CollectorConfig struct {
    AllowedDomains []string
    UserAgents     []string
    Timeout        time.Duration
    RateLimitPerSec int
}

type CollectorBundle struct {
    Collector *colly.Collector
    Queue     *queue.Queue
}

func NewCollector(cfg CollectorConfig) (*CollectorBundle, error) {
    l := logger.New("INFO")
    opts := []func(*colly.Collector){colly.Async(true)}
    if len(cfg.AllowedDomains) > 0 {
        opts = append(opts, colly.AllowedDomains(cfg.AllowedDomains...))
    }
    c := colly.NewCollector(opts...)

    if cfg.Timeout <= 0 {
        cfg.Timeout = 30 * time.Second
    }
    c.SetRequestTimeout(cfg.Timeout)

    if cfg.RateLimitPerSec <= 0 {
        cfg.RateLimitPerSec = 2
    }
    err := c.Limit(&colly.LimitRule{
        DomainGlob:  "*",
        Parallelism: 1,
        RandomDelay: time.Second / time.Duration(cfg.RateLimitPerSec),
    })
    if err != nil {
        return nil, err
    }

    ua := cfg.UserAgents
    if len(ua) == 0 {
        ua = []string{
            "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
            "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
            "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
            "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15A372 Safari/604.1",
            "Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15A372 Safari/604.1",
        }
    }
    c.OnRequest(func(r *colly.Request) {
        r.Headers.Set("User-Agent", ua[rand.Intn(len(ua))])
    })

    c.OnError(func(r *colly.Response, err error) {
        l.Error("scraper_error", map[string]any{"status": r.StatusCode, "url": r.Request.URL.String(), "err": err.Error()})
    })

    c.OnScraped(func(r *colly.Response) {
        l.Info("scraper_done", map[string]any{"url": r.Request.URL.String(), "bytes": len(r.Body)})
    })

    c.OnResponse(func(r *colly.Response) {
        ct := r.Headers.Get("Content-Type")
        if ct != "" {
            if !(strings.Contains(ct, "text/html") || strings.Contains(ct, "application/xhtml+xml")) {
                r.Request.Abort()
            }
        }
    })

    c.WithTransport(&http.Transport{Proxy: http.ProxyFromEnvironment})

    c.IgnoreRobotsTxt = false

    q, err := queue.New(1, &queue.InMemoryQueueStorage{MaxSize: 100})
    if err != nil {
        return nil, err
    }
    return &CollectorBundle{Collector: c, Queue: q}, nil
}

func DefaultCollectorConfig() CollectorConfig {
    var domains []string
    if v := os.Getenv("SCRAPER_ALLOWED_DOMAINS"); v != "" {
        parts := strings.Split(v, ",")
        for _, p := range parts {
            t := strings.TrimSpace(p)
            if t != "" {
                domains = append(domains, t)
            }
        }
    }
    var uas []string
    if v := os.Getenv("SCRAPER_UA_LIST"); v != "" {
        parts := strings.Split(v, "|")
        for _, p := range parts {
            t := strings.TrimSpace(p)
            if t != "" {
                uas = append(uas, t)
            }
        }
    }
    return CollectorConfig{
        AllowedDomains: domains,
        UserAgents:     uas,
        Timeout:        30 * time.Second,
        RateLimitPerSec: 2,
    }
}
