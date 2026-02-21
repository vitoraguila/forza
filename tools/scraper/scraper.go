package scraper

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"github.com/vitoraguila/forza/tools"
)

const (
	DefaultMaxDepth  = 1
	DefaultParallels = 2
	DefaultDelay     = 3
	DefaultAsync     = true

	// Deprecated: Use DefaultMaxDepth instead.
	DefualtMaxDept = DefaultMaxDepth
	// Deprecated: Use DefaultParallels instead.
	DefualtParallels = DefaultParallels
	// Deprecated: Use DefaultDelay instead.
	DefualtDelay = DefaultDelay
	// Deprecated: Use DefaultAsync instead.
	DefualtAsync = DefaultAsync
)

var ErrScrapingFailed = errors.New("scraper could not read URL, or scraping is not allowed for provided URL")

type Scraper struct {
	MaxDepth  int
	Parallels int
	Delay     int64
	Blacklist []string
	Async     bool
}

var _ tools.Tool = Scraper{}

func NewScraper(options ...Options) (*Scraper, error) {
	scraper := &Scraper{
		MaxDepth:  DefaultMaxDepth,
		Parallels: DefaultParallels,
		Delay:     int64(DefaultDelay),
		Async:     DefaultAsync,
		Blacklist: []string{
			"login",
			"signup",
			"signin",
			"register",
			"logout",
			"download",
			"redirect",
		},
	}

	for _, opt := range options {
		opt(scraper)
	}

	return scraper, nil
}

func (s Scraper) Name() string {
	return "web_scraper"
}

func (s Scraper) Description() string {
	return `
		Web Scraper will scan a url and return the content of the web page.
		Input should be a working url.
	`
}

func (s Scraper) Call(ctx context.Context, input string) (string, error) {
	_, err := url.ParseRequestURI(input)
	if err != nil {
		return "", fmt.Errorf("%s: %w", ErrScrapingFailed, err)
	}

	c := colly.NewCollector(
		colly.MaxDepth(s.MaxDepth),
		colly.Async(s.Async),
	)

	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: s.Parallels,
		Delay:       time.Duration(s.Delay) * time.Second,
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", ErrScrapingFailed, err)
	}

	var siteData strings.Builder
	homePageLinks := make(map[string]bool)
	scrapedLinks := make(map[string]bool)
	scrapedLinksMutex := sync.RWMutex{}

	c.OnRequest(func(r *colly.Request) {
		if ctx.Err() != nil {
			r.Abort()
		}
	})

	c.OnHTML("html", func(e *colly.HTMLElement) {
		currentURL := e.Request.URL.String()

		// Only process the page if it hasn't been visited yet
		scrapedLinksMutex.Lock()
		if !scrapedLinks[currentURL] {
			scrapedLinks[currentURL] = true
			scrapedLinksMutex.Unlock()

			siteData.WriteString("\n\nPage URL: " + currentURL)

			title := e.ChildText("title")
			if title != "" {
				siteData.WriteString("\nPage Title: " + title)
			}

			description := e.ChildAttr("meta[name=description]", "content")
			if description != "" {
				siteData.WriteString("\nPage Description: " + description)
			}

			siteData.WriteString("\nHeaders:")
			e.ForEach("h1, h2, h3, h4, h5, h6", func(_ int, el *colly.HTMLElement) {
				siteData.WriteString("\n" + el.Text)
			})

			siteData.WriteString("\nContent:")
			e.ForEach("p", func(_ int, el *colly.HTMLElement) {
				siteData.WriteString("\n" + el.Text)
			})

			if currentURL == input {
				e.ForEach("a", func(_ int, el *colly.HTMLElement) {
					link := el.Attr("href")
					if link != "" && !homePageLinks[link] {
						homePageLinks[link] = true
						siteData.WriteString("\nLink: " + link)
					}
				})
			}
		} else {
			scrapedLinksMutex.Unlock()
		}
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		absoluteLink := e.Request.AbsoluteURL(link)

		// Parse the link to get the hostname
		u, err := url.Parse(absoluteLink)
		if err != nil {
			return
		}

		// Check if the link's hostname matches the current request's hostname
		if u.Hostname() != e.Request.URL.Hostname() {
			return
		}

		// Check for redundant pages
		for _, item := range s.Blacklist {
			if strings.Contains(u.Path, item) {
				return
			}
		}

		// Normalize the path to treat '/' and '/index.html' as the same path
		if u.Path == "/index.html" || u.Path == "" {
			u.Path = "/"
		}

		// Only visit the page if it hasn't been visited yet
		scrapedLinksMutex.RLock()
		if !scrapedLinks[u.String()] {
			scrapedLinksMutex.RUnlock()
			_ = c.Visit(u.String())
		} else {
			scrapedLinksMutex.RUnlock()
		}
	})

	err = c.Visit(input)
	if err != nil {
		return "", fmt.Errorf("%s: %w", ErrScrapingFailed, err)
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		c.Wait()
	}

	// Append all scraped links
	siteData.WriteString("\n\nScraped Links:")
	for link := range scrapedLinks {
		siteData.WriteString("\n" + link)
	}

	return siteData.String(), nil
}
