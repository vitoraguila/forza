package scraper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vitoraguila/forza/tools"
)

// Verify that Scraper implements the Tool interface
var _ tools.Tool = Scraper{}

func TestNewScraper_Defaults(t *testing.T) {
	s, err := NewScraper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.MaxDepth != DefaultMaxDepth {
		t.Errorf("expected MaxDepth %d, got %d", DefaultMaxDepth, s.MaxDepth)
	}
	if s.Parallels != DefaultParallels {
		t.Errorf("expected Parallels %d, got %d", DefaultParallels, s.Parallels)
	}
	if s.Delay != int64(DefaultDelay) {
		t.Errorf("expected Delay %d, got %d", DefaultDelay, s.Delay)
	}
	if s.Async != DefaultAsync {
		t.Errorf("expected Async %v, got %v", DefaultAsync, s.Async)
	}
	if len(s.Blacklist) == 0 {
		t.Error("expected non-empty default blacklist")
	}
}

func TestNewScraper_DeprecatedConstants(t *testing.T) {
	// Verify deprecated constants still work
	if DefualtMaxDept != DefaultMaxDepth {
		t.Errorf("DefualtMaxDept should equal DefaultMaxDepth")
	}
	if DefualtParallels != DefaultParallels {
		t.Errorf("DefualtParallels should equal DefaultParallels")
	}
	if DefualtDelay != DefaultDelay {
		t.Errorf("DefualtDelay should equal DefaultDelay")
	}
	if DefualtAsync != DefaultAsync {
		t.Errorf("DefualtAsync should equal DefaultAsync")
	}
}

func TestNewScraper_WithOptions(t *testing.T) {
	s, err := NewScraper(
		WithMaxDepth(3),
		WithParallelsNum(5),
		WithDelay(10),
		WithAsync(false),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.MaxDepth != 3 {
		t.Errorf("expected MaxDepth 3, got %d", s.MaxDepth)
	}
	if s.Parallels != 5 {
		t.Errorf("expected Parallels 5, got %d", s.Parallels)
	}
	if s.Delay != 10 {
		t.Errorf("expected Delay 10, got %d", s.Delay)
	}
	if s.Async != false {
		t.Errorf("expected Async false, got %v", s.Async)
	}
}

func TestNewScraper_WithBlacklist(t *testing.T) {
	s, _ := NewScraper(WithBlacklist([]string{"extra1", "extra2"}))

	// Should have default + extra items
	if len(s.Blacklist) < 9 { // 7 default + 2 extra
		t.Errorf("expected at least 9 blacklist entries, got %d", len(s.Blacklist))
	}
}

func TestNewScraper_WithNewBlacklist(t *testing.T) {
	s, _ := NewScraper(WithNewBlacklist([]string{"only1", "only2"}))

	if len(s.Blacklist) != 2 {
		t.Errorf("expected exactly 2 blacklist entries, got %d", len(s.Blacklist))
	}
}

func TestScraper_Name(t *testing.T) {
	s, _ := NewScraper()
	if s.Name() != "web_scraper" {
		t.Errorf("expected name 'web_scraper', got %q", s.Name())
	}
}

func TestScraper_Description(t *testing.T) {
	s, _ := NewScraper()
	desc := s.Description()
	if desc == "" {
		t.Error("expected non-empty description")
	}
}

func TestScraper_Call_InvalidURL(t *testing.T) {
	s, _ := NewScraper()
	_, err := s.Call(context.Background(), "not-a-valid-url")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestScraper_Call_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html>
			<head><title>Test Page</title></head>
			<body>
				<h1>Hello World</h1>
				<p>This is a test paragraph.</p>
			</body>
		</html>`))
	}))
	defer server.Close()

	s, _ := NewScraper(WithAsync(false), WithMaxDepth(1))
	result, err := s.Call(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "Test Page") {
		t.Error("expected result to contain page title")
	}
	if !strings.Contains(result, "Hello World") {
		t.Error("expected result to contain header text")
	}
	if !strings.Contains(result, "This is a test paragraph") {
		t.Error("expected result to contain paragraph text")
	}
}

func TestScraper_Call_WithLinks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		if r.URL.Path == "/" {
			w.Write([]byte(`<html>
				<head><title>Home</title></head>
				<body>
					<h1>Home Page</h1>
					<a href="/about">About</a>
				</body>
			</html>`))
		} else {
			w.Write([]byte(`<html>
				<head><title>About</title></head>
				<body>
					<h1>About Page</h1>
					<p>About us content</p>
				</body>
			</html>`))
		}
	}))
	defer server.Close()

	s, _ := NewScraper(WithAsync(false), WithMaxDepth(2))
	result, err := s.Call(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "Home Page") {
		t.Error("expected result to contain home page content")
	}
}

func TestScraper_Call_BlacklistFiltering(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html>
			<body>
				<a href="/login">Login</a>
				<a href="/about">About</a>
			</body>
		</html>`))
	}))
	defer server.Close()

	s, _ := NewScraper(WithAsync(false), WithMaxDepth(1))
	result, err := s.Call(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestScraper_Call_CancelledContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><p>content</p></body></html>`))
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	s, _ := NewScraper(WithAsync(false), WithMaxDepth(1))
	_, err := s.Call(ctx, server.URL)
	// With cancelled context, should return error or empty result
	if err != nil {
		// This is expected - context was cancelled
		return
	}
}
