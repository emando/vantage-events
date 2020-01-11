// Copyright © 2020 Emando B.V.

package entities

// Competition is a Vantage competition.
type Competition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Distance is a Vantage competition distance.
type Distance struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Heat is a Vantage competition distance heat.
type Heat struct {
	Key struct {
		Round  int `json:"round"`
		Number int `json:"number"`
	} `json:"heat"`
}
