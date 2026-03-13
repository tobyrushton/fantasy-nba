// Package scraper is responsible for obtaining player and game data.
package scraper

import "github.com/tobyrushton/fantasy-nba/pkg/client"

// Our scraper implementation that is in charge of obtaining data.
type Scraper struct {
	// The HTTP client used to make requests.
	client client.Client
}

// Initialises a new scraper instance.
func New(client client.Client) *Scraper {
	return &Scraper{
		client: client,
	}
}
