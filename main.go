package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"time"
)

type Scraper struct {
	ctx        context.Context
	client     *http.Client
	f          *os.File
	outName    string
	wroteFirst bool
	wroteCount int64
}

func NewScraper(ctx context.Context, outName string, f *os.File) *Scraper {
	return &Scraper{
		ctx:     ctx,
		client:  http.DefaultClient,
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
		res, err := s.getPage(s.ctx, page)
		if err != nil {
			log.Printf("Fetch page %d error: %v", page, err)
			return
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

		time.Sleep(rand.N(100 * time.Millisecond))
	}
}

type Deal struct {
	InternalName       string `json:"internalName"`
	Title              string `json:"title"`
	MetacriticLink     string `json:"metacriticLink"`
	DealID             string `json:"dealID"`
	StoreID            string `json:"storeID"`
	GameID             string `json:"gameID"`
	SalePrice          string `json:"salePrice"`
	NormalPrice        string `json:"normalPrice"`
	IsOnSale           string `json:"isOnSale"`
	Savings            string `json:"savings"`
	MetacriticScore    string `json:"metacriticScore"`
	SteamRatingText    string `json:"steamRatingText"`
	SteamRatingPercent string `json:"steamRatingPercent"`
	SteamRatingCount   string `json:"steamRatingCount"`
	SteamAppID         string `json:"steamAppID"`
	ReleaseDate        int64  `json:"releaseDate"`
	LastChange         int64  `json:"lastChange"`
	DealRating         string `json:"dealRating"`
	Thumb              string `json:"thumb"`
}

func (s *Scraper) getPage(ctx context.Context, page uint) ([]Deal, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.cheapshark.com/api/1.0/deals", nil)
	if err != nil {
		return nil, err
	}

	req.URL.Query().Set("storeID", "1")
	req.URL.Query().Set("page", fmt.Sprint(page))
	req.URL.Query().Set("sortBy", "Release")

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var deals []Deal
	if err = json.NewDecoder(res.Body).Decode(&deals); err != nil {
		return nil, err
	}

	return deals, nil
}

func main() {
	outName := fmt.Sprintf("cheapshark-%s.json", time.Now().UTC().Format("20060102"))
	f, err := os.Create(outName)
	if err != nil {
		log.Fatalf("create output file: %v", err)
	}
	defer f.Close()

	s := NewScraper(context.Background(), outName, f)
	s.writeHeader()
	s.processAllPages()
	s.writeFooter()

	log.Printf("Scraping completed. Result written to %s", s.outName)
	log.Printf("Actually written=%d", s.wroteCount)
}
