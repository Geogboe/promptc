package fetcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCacheKey_URL(t *testing.T) {
	r := Resource{URL: "https://example.com/docs", Name: "docs"}
	key := CacheKey(r)
	if len(key) != 64 {
		t.Errorf("expected 64-char hex SHA-256, got len=%d", len(key))
	}
	// Same input → same key
	if CacheKey(r) != key {
		t.Error("CacheKey is not deterministic")
	}
}

func TestCacheKey_Git_DefaultRef(t *testing.T) {
	r1 := Resource{Git: "https://github.com/org/repo", Name: "repo"}
	r2 := Resource{Git: "https://github.com/org/repo", Name: "repo", Ref: "HEAD"}
	// ref="" and ref="HEAD" should produce the same key
	if CacheKey(r1) != CacheKey(r2) {
		t.Error("empty ref and HEAD should produce same cache key")
	}
}

func TestCacheKey_DifferentInputs(t *testing.T) {
	r1 := Resource{URL: "https://example.com/a", Name: "a"}
	r2 := Resource{URL: "https://example.com/b", Name: "b"}
	if CacheKey(r1) == CacheKey(r2) {
		t.Error("different URLs should produce different cache keys")
	}
}

func TestFetchHTTP_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hello from test server"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	r := Resource{URL: srv.URL, Name: "test-resource"}
	if err := FetchHTTP(context.Background(), r, dir); err != nil {
		t.Fatalf("FetchHTTP failed: %v", err)
	}

	outPath := filepath.Join(dir, "test-resource.md")
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	if !strings.Contains(string(data), "Hello from test server") {
		t.Error("output missing expected content")
	}
}

func TestFetchHTTP_HTMLStripping(t *testing.T) {
	html := `<html><head><title>Test</title></head>
<body>
<h1>Hello World</h1>
<p>This is a paragraph.</p>
<script>console.log("ignored")</script>
</body></html>`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}))
	defer srv.Close()

	dir := t.TempDir()
	r := Resource{URL: srv.URL, Name: "html-resource"}
	if err := FetchHTTP(context.Background(), r, dir); err != nil {
		t.Fatalf("FetchHTTP failed: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "html-resource.md"))
	content := string(data)

	if strings.Contains(content, "<html>") || strings.Contains(content, "<p>") {
		t.Error("HTML tags should be stripped from output")
	}
	if strings.Contains(content, "console.log") {
		t.Error("script content should be stripped")
	}
	if !strings.Contains(content, "Hello World") {
		t.Error("text content should be preserved")
	}
	if !strings.Contains(content, "This is a paragraph") {
		t.Error("paragraph text should be preserved")
	}
}

func TestFetchHTTP_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	dir := t.TempDir()
	r := Resource{URL: srv.URL, Name: "missing"}
	if err := FetchHTTP(context.Background(), r, dir); err == nil {
		t.Error("expected error for HTTP 404")
	}
}

func TestHtmlToText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
		excludes string
	}{
		{
			name:     "strips tags",
			input:    "<p>Hello <b>World</b></p>",
			contains: "Hello World",
			excludes: "<p>",
		},
		{
			name:     "decodes entities",
			input:    "AT&amp;T &lt;tag&gt;",
			contains: "AT&T <tag>",
		},
		{
			name:     "removes scripts",
			input:    "<script>alert('xss')</script>content",
			contains: "content",
			excludes: "alert",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := htmlToText(tt.input)
			if tt.contains != "" && !strings.Contains(result, tt.contains) {
				t.Errorf("expected %q in result: %q", tt.contains, result)
			}
			if tt.excludes != "" && strings.Contains(result, tt.excludes) {
				t.Errorf("expected %q NOT in result: %q", tt.excludes, result)
			}
		})
	}
}

func TestFetchAll_Empty(t *testing.T) {
	errs := FetchAll(context.Background(), nil, t.TempDir())
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty resource list, got: %v", errs)
	}
}

func TestFetchAll_MultipleHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("content for " + r.URL.Path))
	}))
	defer srv.Close()

	dir := t.TempDir()
	resources := []Resource{
		{URL: srv.URL + "/a", Name: "resource-a"},
		{URL: srv.URL + "/b", Name: "resource-b"},
		{URL: srv.URL + "/c", Name: "resource-c"},
	}

	errs := FetchAll(context.Background(), resources, dir)
	if len(errs) != 0 {
		t.Errorf("unexpected errors: %v", errs)
	}

	for _, r := range resources {
		path := filepath.Join(dir, r.Name+".md")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("output file not created: %s", path)
		}
	}
}

func TestFetchAll_PartialFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/fail" {
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	resources := []Resource{
		{URL: srv.URL + "/ok", Name: "ok-resource"},
		{URL: srv.URL + "/fail", Name: "fail-resource"},
	}

	errs := FetchAll(context.Background(), resources, dir)
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errs), errs)
	}
	// The successful resource should still exist
	if _, err := os.Stat(filepath.Join(dir, "ok-resource.md")); os.IsNotExist(err) {
		t.Error("successful resource should be written even when others fail")
	}
}
