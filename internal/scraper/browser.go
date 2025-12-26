package scraper

import (
	"context"
	"errors"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"fmt"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type BrowserPool struct {
	browsers []*rod.Browser
	mu       sync.Mutex
	idx      int
	sem      chan struct{}
	idleTTL  time.Duration
	lastUse  time.Time
}

func NewBrowserPool(size int, idleTTL time.Duration) (*BrowserPool, error) {
	if size <= 0 {
		size = 2
	}
	bp := &BrowserPool{sem: make(chan struct{}, size), idleTTL: idleTTL}
	for i := 0; i < size; i++ {
		u := launcher.New().Headless(true).MustLaunch()
		b := rod.New().ControlURL(u).MustConnect()
		bp.browsers = append(bp.browsers, b)
	}
	bp.lastUse = time.Now()
	go bp.reaper()
	return bp, nil
}

func (bp *BrowserPool) reaper() {
	for {
		time.Sleep(time.Second)
		if time.Since(bp.lastUse) > bp.idleTTL {
			for _, b := range bp.browsers {
				_ = b.Close()
			}
			bp.browsers = nil
			return
		}
	}
}

func (bp *BrowserPool) acquire() { bp.sem <- struct{}{} }
func (bp *BrowserPool) release() { <-bp.sem }

func (bp *BrowserPool) nextBrowser() *rod.Browser {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.lastUse = time.Now()
	if len(bp.browsers) == 0 {
		u := launcher.New().Headless(true).MustLaunch()
		b := rod.New().ControlURL(u).MustConnect()
		bp.browsers = append(bp.browsers, b)
		bp.idx = 0
		return b
	}
	b := bp.browsers[bp.idx%len(bp.browsers)]
	bp.idx++
	return b
}

type DynamicConfig struct {
	PoolSize   int
	IdleTTL    time.Duration
	NavTimeout time.Duration
	Allowed    []string
}

func DefaultDynamicConfig() DynamicConfig {
	ps := 2
	if v := os.Getenv("SCRAPER_BROWSER_POOL_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			ps = n
		}
	}
	idle := 60 * time.Second
	if v := os.Getenv("SCRAPER_DYNAMIC_IDLE_SECS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			idle = time.Duration(n) * time.Second
		}
	}
	nt := 10 * time.Second
	if v := os.Getenv("SCRAPER_NAV_TIMEOUT_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			nt = time.Duration(n) * time.Millisecond
		}
	}
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
	return DynamicConfig{PoolSize: ps, IdleTTL: idle, NavTimeout: nt, Allowed: domains}
}

func allowed(urlStr string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	u, err := url.Parse(urlStr)
	if err != nil || u.Host == "" {
		return false
	}
	host := u.Host
	if h, p, err := net.SplitHostPort(u.Host); err == nil && h != "" {
		host = h
		_ = p
	}
	h := strings.ToLower(host)
	for _, d := range allowed {
		d = strings.ToLower(strings.TrimSpace(d))
		if h == d || strings.HasSuffix(h, "."+d) {
			return true
		}
	}
	return false
}

func ScrapeDynamicURL(urlStr string, cfg DynamicConfig) (string, error) {
	if !allowed(urlStr, cfg.Allowed) {
		return "", errors.New("domain_not_allowed")
	}
	bp, err := NewBrowserPool(cfg.PoolSize, cfg.IdleTTL)
	if err != nil {
		return "", fmt.Errorf("new browser pool: %w", err)
	}
	bp.acquire()
	defer bp.release()
	br := bp.nextBrowser()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.NavTimeout)
	defer cancel()
	br = br.Context(ctx)
	p, err := br.Page(proto.TargetCreateTarget{URL: urlStr})
	if err != nil {
		return "", fmt.Errorf("open page: %w", err)
	}
	defer func() { _ = p.Close() }()
	_ = p.WaitLoad()
	time.Sleep(3 * time.Second)

	html, err := p.HTML()
	if err != nil {
		return "", fmt.Errorf("get html: %w", err)
	}
	if len(html) > 2<<20 {
		html = html[:2<<20]
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", fmt.Errorf("new goquery doc: %w", err)
	}
	txt := visibleText(doc.Find("body"))
	n, _ := NormalizeText(txt)
	return strings.TrimSpace(n), nil
}
