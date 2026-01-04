//go:build fitz

package pdf

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gen2brain/go-fitz"
	"github.com/onurceri/botla-app/internal/scraper"
	hn "golang.org/x/net/html"
)

func ExtractPDFText(filePath string, _ string, _ bool) (string, error) {
	doc, err := fitz.New(filePath)
	if err != nil {
		return "", &PDFError{Op: "open", Err: err}
	}
	defer doc.Close()

	pages := doc.NumPage()
	if pages < 1 {
		return "", &PDFError{Op: "validate", Err: fmt.Errorf("no pages")}
	}

	var out strings.Builder
	for n := 0; n < pages; n++ {
		// prefer HTML-based positional extraction to preserve layout
		pageText, err := extractPageTextHTML(doc, n)
		if err != nil || strings.TrimSpace(pageText) == "" {
			// fallback to simple text extraction
			t, terr := doc.Text(n)
			if terr != nil {
				return "", &PDFError{Op: fmt.Sprintf("extract_text_page_%d", n), Err: terr}
			}
			pageText = t
		}

		if !scraper.IsValidUTF8([]byte(pageText)) {
			pageText = string(bytes.ToValidUTF8([]byte(pageText), []byte("?")))
		}
		norm, nerr := scraper.NormalizeText(pageText)
		if nerr != nil {
			norm = strings.TrimSpace(pageText)
		}
		if norm != "" {
			out.WriteString(norm)
		}
		if n < pages-1 {
			out.WriteString("\n\n")
		}
	}
	res := strings.TrimSpace(out.String())
	return res, nil
}

type span struct {
	x float64
	y float64
	t string
}

var (
	reLeft = regexp.MustCompile(`left:\s*([0-9]+\.?[0-9]*)`)
	reTop  = regexp.MustCompile(`top:\s*([0-9]+\.?[0-9]*)`)
)

func extractPageTextHTML(doc *fitz.Document, page int) (string, error) {
	html, err := doc.HTML(page, false)
	if err != nil {
		return "", err
	}
	n, err := hn.Parse(strings.NewReader(html))
	if err != nil {
		return "", err
	}
	var spans []span
	var walk func(*hn.Node)
	walk = func(node *hn.Node) {
		if node.Type == hn.ElementNode && node.Data == "span" {
			var style string
			for _, a := range node.Attr {
				if a.Key == "style" {
					style = a.Val
					break
				}
			}
			var buf strings.Builder
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == hn.TextNode {
					buf.WriteString(c.Data)
				}
			}
			txt := strings.TrimSpace(buf.String())
			if txt != "" && style != "" {
				x := parseCoord(reLeft, style)
				y := parseCoord(reTop, style)
				spans = append(spans, span{x: x, y: y, t: txt})
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)

	if len(spans) == 0 {
		return "", nil
	}
	sort.Slice(spans, func(i, j int) bool {
		if spans[i].y == spans[j].y {
			return spans[i].x < spans[j].x
		}
		return spans[i].y < spans[j].y
	})

	const tol = 2.0
	var out strings.Builder
	var lineY float64 = -1
	first := true
	for _, s := range spans {
		if lineY < 0 || abs(s.y-lineY) > tol {
			if !first {
				out.WriteString("\n")
			}
			first = false
			lineY = s.y
			out.WriteString(s.t)
		} else {
			out.WriteString(" ")
			out.WriteString(s.t)
		}
	}
	return out.String(), nil
}

func parseCoord(re *regexp.Regexp, style string) float64 {
	m := re.FindStringSubmatch(style)
	if len(m) == 2 {
		// best-effort parse; ignore error -> 0
		v, _ := strconv.ParseFloat(m[1], 64)
		return v
	}
	return 0
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
