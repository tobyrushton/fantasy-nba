package client

import "net/http"

//go:generate go tool counterfeiter -generate

// Client is an interface that abstracts the HTTP client used to make requests.
//
//counterfeiter:generate -o ../fakes/client.go . Client
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}
