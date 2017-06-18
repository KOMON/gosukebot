package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Card struct {
	ID            string   `json:"id"`
	Layout        string   `json:"layout"`
	Name          string   `json:"name"`
	ManaCost      string   `json:"manaCost"`
	CMC           float64  `json:"cmc"`
	Colors        []string `json:"colors"`
	ColorIdentity []string `json:"colorIdentity"`
	Type          string   `json:"type"`
	Supertypes    []string `json:"supertypes"`
	Types         []string `json:"types"`
	Subtypes      []string `json:"subtypes"`
	Rarity        Rarity   `json:"rarity"`
	Text          string   `json:"text"`
	Flavor        string   `json:"flavor"`
	Artist        string   `json:"artist"`
	Number        string   `json:"number"`
	Power         string   `json:"power"`
	Toughness     string   `json:"toughness"`
	Loyalty       float64 `json:"loyalty"`
	MultiverseID  float64 `json:"multiverseid"`
	Timeshifted   bool    `json:"timeshifted"`
	Reserved      bool    `json:"reserved"`
	ReleaseDate   string  `json:"releaseDate"`
	MCINumber     string  `json:"mciNumber"`
}

type Set struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	ReleaseDate string `json:"releaseDate"`
	Type        string `json:"type"`
	Block       string `json:"block"`
	Cards       []Card `json:"cards"`
}

type Sets map[string]Set

type Cards map[string]Card

func loadSets(filename string) (Sets, error) {
	var sets Sets

	body, err := readJSON(filename)
	if err != nil {
		return Sets{}, err
	}
	err = json.Unmarshal(body, &sets)
	if err != nil {
		return Sets{}, err
	}

	return sets, err
}

func loadCards(filename string) (Cards, error) {
	var cards Cards

	body, err := readJSON(filename)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &cards)
	if err != nil {
		return nil, err
	}

	return cards, err
}

func readJSON(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

type Rarity struct {
	Rarities []string
}

func (r *Rarity) UnmarshalJSON(b []byte) (err error) {
	slice, str := []string{}, ""

	if err = json.Unmarshal(b, &str); err == nil {
		r.Rarities = append(r.Rarities, str)
		return
	}

	if err = json.Unmarshal(b, &slice); err == nil {
		for _, s := range slice {
			r.Rarities = append(r.Rarities, s)
		}
		return
	}

	return
}

func (r Rarity) String() string {
	if len(r.Rarities) == 0 {
		return ""
	}

	if len(r.Rarities) == 1 {
		return r.Rarities[0]
	}

	s := "["
	for _, str := range r.Rarities {
		s = s + str + ","
	}
	return s + "]"
}
