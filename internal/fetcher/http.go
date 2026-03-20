package fetcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	getter "github.com/hashicorp/go-getter"
)

// FetchHTTP downloads a URL using go-getter and saves the result as
// <destDir>/<resource.Name>.md. HTML responses are converted to readable text.
func FetchHTTP(ctx context.Context, resource Resource, destDir string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", destDir, err)
	}

	tmpFile := filepath.Join(destDir, resource.Name+".tmp")
	outFile := filepath.Join(destDir, resource.Name+".md")

	client := &getter.Client{
		Ctx:  ctx,
		Src:  resource.URL,
		Dst:  tmpFile,
		Mode: getter.ClientModeFile,
	}
	if err := client.Get(); err != nil {
		return fmt.Errorf("downloading %s: %w", resource.URL, err)
	}

	raw, err := os.ReadFile(tmpFile)
	_ = os.Remove(tmpFile)
	if err != nil {
		return fmt.Errorf("reading downloaded file: %w", err)
	}

	content := string(raw)
	if looksLikeHTML(content) {
		content = htmlToText(content)
	}

	return os.WriteFile(outFile, []byte(content), 0644)
}

func looksLikeHTML(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(strings.ToLower(s), "<!doctype") ||
		strings.HasPrefix(strings.ToLower(s), "<html")
}

// htmlToText strips HTML tags and decodes common entities to produce
// readable plain text. Not a full HTML parser — sufficient for docs pages.
var (
	tagRe      = regexp.MustCompile(`<[^>]+>`)
	whitespace = regexp.MustCompile(`[ \t]+`)
	blankLines = regexp.MustCompile(`\n{3,}`)
)

func htmlToText(html string) string {
	html = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`).ReplaceAllString(html, "")

	for _, tag := range []string{"p", "div", "br", "li", "h1", "h2", "h3", "h4", "h5", "h6", "tr"} {
		html = regexp.MustCompile(`(?i)</?`+tag+`[^>]*>`).ReplaceAllString(html, "\n")
	}

	html = tagRe.ReplaceAllString(html, "")
	html = strings.NewReplacer(
		"&amp;", "&", "&lt;", "<", "&gt;", ">",
		"&quot;", `"`, "&#39;", "'", "&nbsp;", " ",
	).Replace(html)

	lines := strings.Split(html, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = whitespace.ReplaceAllString(line, " ")
		out = append(out, strings.TrimSpace(line))
	}
	result := strings.Join(out, "\n")
	result = blankLines.ReplaceAllString(result, "\n\n")
	return strings.TrimSpace(result)
}
