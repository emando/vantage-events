// Copyright © 2019 Emando B.V.

package events

type Competition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Distance struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Heat struct {
	Key struct {
		Round  int `json:"round"`
		Number int `json:"number"`
	} `json:"heat"`
}
