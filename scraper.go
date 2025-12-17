package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"golang.org/x/time/rate"
)

type Scraper struct {
	ctx        context.Context
	client     *http.Client
	limiter    *rate.Limiter
	f          *os.File
	outName    string
	wroteFirst bool
	wroteCount int64
}

func NewScraper(ctx context.Context, outName string, f *os.File, limit rate.Limit, burst int) *Scraper {
	return &Scraper{
		ctx:     ctx,
		client:  http.DefaultClient,
		limiter: rate.NewLimiter(limit, burst),
		f:       f,
		outName: outName,
	}
}

func (s *Scraper) writeHeader() {
	header := struct {
		SchemaVersion int    `json:"schema_version"`
		ScrapedAt     string `json:"scraped_at"`
		Source        string `json:"source"`
	}{
		SchemaVersion: 1,
		ScrapedAt:     time.Now().UTC().Format(time.RFC3339),
		Source:        "cheapshark",
	}
	encHeader, _ := json.Marshal(header)
	if _, err := s.f.Write(encHeader[:len(encHeader)-1]); err != nil {
		log.Fatalf("write header: %v", err)
	}
	if _, err := s.f.WriteString(",\"data\":["); err != nil {
		log.Fatalf("write data open bracket: %v", err)
	}
}

func (s *Scraper) writeFooter() {
	if _, err := s.f.WriteString("]}"); err != nil {
		log.Fatalf("write footer: %v", err)
	}
}

func (s *Scraper) writeItemBytes(b []byte) {
	if !s.wroteFirst {
		if _, err := s.f.Write(b); err != nil {
			log.Printf("Write first item error: %v", err)
			return
		}
		s.wroteFirst = true
	} else {
		if _, err := s.f.WriteString(","); err != nil {
			log.Printf("Write comma error: %v", err)
			return
		}
		if _, err := s.f.Write(b); err != nil {
			log.Printf("Write item error: %v", err)
			return
		}
	}
	s.wroteCount++
}

func (s *Scraper) processAllPages() {
	page := uint(0)
	for {
		res, totalPages, err := s.getPage(s.ctx, page)
		if err != nil {
			log.Printf("Fetch page %d error: %v", page, err)
			return
		}
		if page > totalPages {
			break
		}
		for _, deal := range res {
			b, err := json.Marshal(deal)
			if err != nil {
				log.Printf("Marshal deal error: %v", err)
				continue
			}
			s.writeItemBytes(b)
		}
		log.Printf("Page %d scraped", page)
		page++

		if err = s.limiter.Wait(s.ctx); err != nil {
			log.Printf("Wait error: %v", err)
			return
		}
	}
}

func (s *Scraper) getPage(ctx context.Context, page uint) ([]Deal, uint, error) {
	u, err := url.Parse("https://www.cheapshark.com/api/1.0/deals")
	if err != nil {
		return nil, 0, err
	}

	q := u.Query()
	q.Set("storeID", "1")
	q.Set("page", fmt.Sprint(page))
	q.Set("sortBy", "Release")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, 0, err
	}

	res, err := s.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, 0, fmt.Errorf("bad status code: %d", res.StatusCode)
		}

		return nil, 0, fmt.Errorf("bad status code: %d, response: %s", res.StatusCode, string(body))
	}

	var deals []Deal
	if err = json.NewDecoder(res.Body).Decode(&deals); err != nil {
		return nil, 0, err
	}

	totalPages, err := strconv.ParseUint(res.Header.Get("x-total-page-count"), 10, 64)
	if err != nil {
		return nil, 0, err
	}

	return deals, uint(totalPages), nil
}
