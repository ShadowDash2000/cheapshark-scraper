package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

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
