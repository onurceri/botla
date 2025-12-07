package scraper

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/onurceri/botla-co/pkg/logger"
	"golang.org/x/net/html"
)

type ScrapingTask struct {
	URL       string
	ChatbotID int64
	SourceID  int64
}

func keyFor(url string) string {
	h := sha256.Sum256([]byte(url))
	return "scraped:" + hex.EncodeToString(h[:])
}

func visibleText(sel *goquery.Selection) string {
	sel.Find("script,style,noscript").Remove()
	sel.Find(`[style*="display:none"]`).Remove()
	sel.Find(`[hidden]`).Remove()
	sel.Find(`.hidden`).Remove()
	sel.Find(`[aria-hidden="true"]`).Remove()

	// Inject separator for block elements to preserve structure
	sel.Find("br, p, div, li, h1, h2, h3, h4, h5, h6, tr, article, section, header, footer").PrependHtml(" ||BLOCK|| ")

	t := strings.TrimSpace(sel.Text())
	// Collapse whitespace
	t = strings.Join(strings.Fields(t), " ")
	// Restore newlines
	t = strings.ReplaceAll(t, "||BLOCK||", "\n")
	return t
}

func ScrapeURL(task ScrapingTask, cfg CollectorConfig) (string, error) {
	l := logger.New("INFO")
	cache := NewCache()
	k := keyFor(task.URL)
	if v, ok := cache.Get(k); ok {
		return v, nil
	}

	bundle, err := NewCollector(cfg)
	if err != nil {
		return "", err
	}
	c := bundle.Collector
	var content string

	c.OnHTML("body", func(e *colly.HTMLElement) {
		sel := e.DOM
		txt := visibleText(sel)
		n, _ := NormalizeText(txt)
		content = strings.Join(strings.Fields(n), " ")
	})

	err = bundle.Queue.AddURL(task.URL)
	if err != nil {
		return "", err
	}
	if err := bundle.Queue.Run(c); err != nil {
		return "", err
	}
	c.Wait()

	if content == "" {
		l.Warn("scraper_empty", map[string]any{"url": task.URL})
		return "", nil
	}

	_ = cache.Set(k, content, 7*24*time.Hour)
	l.Info("scraper_cached", map[string]any{"url": task.URL, "len": len(content)})
	return content, nil
}

// ExtractLinks finds all links in the HTML content that belong to the same domain as baseURL.
// It returns a list of absolute URLs.
func ExtractLinks(htmlContent string, baseURL string) ([]string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	var links []string
	seen := make(map[string]bool)

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					val := strings.TrimSpace(a.Val)
					if val == "" || strings.HasPrefix(val, "#") || strings.HasPrefix(val, "javascript:") || strings.HasPrefix(val, "mailto:") {
						continue
					}
					u, err := base.Parse(val)
					if err == nil {
						// Normalize: remove fragment, query? maybe keep query for some sites?
						// Let's remove fragment for sure.
						u.Fragment = ""
						// Only internal links (subdomains included? usually yes or strict domain match)
						// Let's do strict host match for now or subdomain allowed?
						// Usually "same domain" means everything under example.com including sub.example.com
						// But safer is same host.
						if u.Host == base.Host {
							s := u.String()
							if !seen[s] && s != baseURL {
								seen[s] = true
								links = append(links, s)
							}
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return links, nil
}

func ScrapeURLWithFallback(task ScrapingTask, cfg CollectorConfig, allowDynamic bool) (string, error) {
	l := logger.New("INFO")
	cache := NewCache()
	k := keyFor(task.URL)
	if v, ok := cache.Get(k); ok {
		return v, nil
	}
	// First try static
	s, err := ScrapeURL(task, cfg)
	if err == nil && s != "" {
		if len(s) > 1000 {
			_ = cache.Set(k, s, 7*24*time.Hour)
			return s, nil
		}
		// If static content is short, try dynamic to see if we get more
		l.Info("scraper_static_short", map[string]any{"url": task.URL, "len": len(s)})
	}

	if !allowDynamic {
		if s != "" {
			_ = cache.Set(k, s, 7*24*time.Hour)
			return s, nil
		}
		if err != nil {
			return "", err
		}
		// If s is empty and no error (e.g. empty page), return empty
		return "", nil
	}

	// Fallback to dynamic
	dc := DefaultDynamicConfig()
	ds, derr := ScrapeDynamicURL(task.URL, dc)
	if derr == nil && ds != "" {
		_ = cache.Set(k, ds, 7*24*time.Hour)
		l.Info("scraper_dynamic_ok", map[string]any{"url": task.URL, "len": len(ds)})
		return ds, nil
	}
	if err != nil {
		l.Warn("scraper_static_fail", map[string]any{"url": task.URL, "err": err.Error()})
	}
	if derr != nil {
		l.Warn("scraper_dynamic_fail", map[string]any{"url": task.URL, "err": derr.Error()})
	}
	if s != "" {
		return s, nil
	}
	return "", nil
}
