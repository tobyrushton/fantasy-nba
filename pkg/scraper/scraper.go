// Package scraper is responsible for obtaining player and game data.
package scraper

import "net/http"

// Client is an interface that abstracts the HTTP client used to make requests.
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// Our scraper implementation that is in charge of obtaining data.
type Scraper struct {
	// The HTTP client used to make requests.
	client Client
}

// Initialises a new scraper instance.
func New(client Client) *Scraper {
	return &Scraper{
		client: client,
	}
}
