package main

import (
	"encoding/json"
	"os"
	"testing"
)

func Test_DealUnmarshal(t *testing.T) {
	f, err := os.Open("test_files/deal.json")
	if err != nil {
		t.Fatal(err)
	}

	var d Deal
	err = json.NewDecoder(f).Decode(&d)
	if err != nil {
		t.Fatal(err)
	}

	if d.GameID != 303069 {
		t.Errorf("GameID is %d, expected 303069", d.GameID)
	}
}

func Test_DealsUnmarshal(t *testing.T) {
	f, err := os.Open("test_files/deals.json")
	if err != nil {
		t.Fatal(err)
	}

	var ds []Deal
	err = json.NewDecoder(f).Decode(&ds)
	if err != nil {
		t.Fatal(err)
	}
}
