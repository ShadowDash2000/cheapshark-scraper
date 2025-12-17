package main

import (
	"bytes"
	"strconv"
)

type Deal struct {
	InternalName       string      `json:"internalName"`
	Title              string      `json:"title"`
	MetacriticLink     string      `json:"metacriticLink"`
	DealID             string      `json:"dealID"`
	StoreID            JsonUint    `json:"storeID"`
	GameID             JsonUint    `json:"gameID"`
	SalePrice          JsonFloat64 `json:"salePrice"`
	NormalPrice        JsonFloat64 `json:"normalPrice"`
	IsOnSale           JsonBool    `json:"isOnSale"`
	Savings            JsonFloat64 `json:"savings"`
	MetacriticScore    JsonUint    `json:"metacriticScore"`
	SteamRatingText    string      `json:"steamRatingText"`
	SteamRatingPercent JsonUint    `json:"steamRatingPercent"`
	SteamRatingCount   JsonUint    `json:"steamRatingCount"`
	SteamAppID         JsonUint    `json:"steamAppID"`
	ReleaseDate        int64       `json:"releaseDate"`
	LastChange         int64       `json:"lastChange"`
	DealRating         JsonFloat64 `json:"dealRating"`
	Thumb              string      `json:"thumb"`
}

type JsonBool bool

func (b *JsonBool) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("\"1\"")) {
		*b = true
	}

	*b = false

	return nil
}

type JsonFloat64 float64

func (f *JsonFloat64) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*f = 0
		return nil
	}

	str, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}

	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return err
	}

	*f = JsonFloat64(val)
	return nil
}

type JsonUint uint

func (u *JsonUint) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*u = 0
		return nil
	}

	str, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}

	val, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return err
	}

	*u = JsonUint(val)
	return nil
}
