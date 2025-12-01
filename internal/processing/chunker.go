package processing

import (
    "strings"
    "unicode"
)

// ChunkText splits input into overlapping chunks preserving paragraph/word boundaries.
// targetLen is the desired maximum chunk length; overlap is appended from previous chunk end.
func ChunkText(input string, targetLen int, overlap int) []string {
    s := strings.TrimSpace(input)
    if s == "" || targetLen <= 0 {
        return nil
    }
    paras := splitParagraphs(s)
    var chunks []string
    var buf strings.Builder
    var lastTail string

    write := func() {
        if buf.Len() == 0 {
            return
        }
        out := strings.TrimSpace(buf.String())
        if out != "" {
            if overlap > 0 && lastTail != "" {
                out = strings.TrimSpace(lastTail) + "\n" + out
            }
            chunks = append(chunks, out)
            // compute new tail
            tail := out
            if len(tail) > overlap {
                tail = tail[len(tail)-overlap:]
                // trim to word boundary
                i := strings.IndexFunc(tail, unicode.IsSpace)
                if i > 0 {
                    tail = tail[i:]
                }
            }
            lastTail = tail
        }
        buf.Reset()
    }

    // Pre-process paragraphs to split oversized ones
    var refinedParas []string
    for _, p := range paras {
        p = strings.TrimSpace(p)
        if len(p) > targetLen {
            for len(p) > targetLen {
                cut := targetLen
                // Backtrack to find a space to avoid cutting words
                found := false
                startSearch := cut
                if startSearch >= len(p) { startSearch = len(p) - 1 }
                
                for i := startSearch; i > cut-200 && i > 0; i-- {
                    if unicode.IsSpace(rune(p[i])) {
                        cut = i
                        found = true
                        break
                    }
                }
                if !found {
                    // No space found, force cut
                    cut = targetLen
                }
                
                refinedParas = append(refinedParas, p[:cut])
                if cut+1 < len(p) {
                    p = strings.TrimSpace(p[cut:])
                } else {
                    p = ""
                }
            }
            if len(p) > 0 {
                refinedParas = append(refinedParas, p)
            }
        } else {
            refinedParas = append(refinedParas, p)
        }
    }
    paras = refinedParas

    for _, p := range paras {
        p = strings.TrimSpace(p)
        if p == "" {
            continue
        }
        if buf.Len()+len(p)+1 > targetLen {
            write()
        }
        if buf.Len() > 0 {
            buf.WriteString("\n\n")
        }
        buf.WriteString(p)
        if buf.Len() >= targetLen {
            write()
        }
    }
    write()
    return chunks
}

func splitParagraphs(s string) []string {
    // Split by double newlines first, then fallback to single newline
    parts := strings.Split(s, "\n\n")
    if len(parts) == 1 {
        parts = strings.Split(s, "\n")
    }
    // coarse sentence join: avoid micro-chunks
    var out []string
    for _, p := range parts {
        p = strings.TrimSpace(p)
        if p != "" {
            out = append(out, p)
        }
    }
    return out
}

