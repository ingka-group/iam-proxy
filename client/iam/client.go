// Copyright © 2024 Ingka Holding B.V. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package iam

//go:generate gowrap gen -g -i Servicer -t ../../templates/opencensus_metrics.tpl -o client_with_metrics_gen.go

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/ingka-group-digital/iam-proxy/client/health"
	"github.com/ingka-group-digital/iam-proxy/client/iamerrors"
	"github.com/ingka-group-digital/iam-proxy/client/paths"
)

const (
	// ServiceName is the name of the service
	ServiceName = "iam-proxy"
)

var (
	// ErrInvalidDomain is the error when the domain provided is invalid.
	ErrInvalidDomain = errors.New("invalid domain")
	// ErrInvalidPort is the error when the port provided is invalid.
	ErrInvalidPort = errors.New("invalid port")
)

// Servicer defines the interface for interacting with the server.
type Servicer interface {
	Health() (*health.Health, error)
	Ready() error
	Token() (string, error)
	Validate(token string) error
	Identity(token string) (string, error)
}

// Client implements the service logic by making http requests to the server.
type Client struct {
	URL        string
	HTTPClient *http.Client
}

// New returns a client
func New(url string, client *http.Client) *Client {
	return &Client{
		URL:        url,
		HTTPClient: client,
	}
}

// NewDefault returns a client with default configuration
func NewDefault() *Client {
	return &Client{
		URL: "http://" + ServiceName,
		// Wrap HTTP Client with otel to propagate traces
		HTTPClient: &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)},
	}
}

// Health calls the health endpoint.
func (c *Client) Health() (*health.Health, error) {
	url := c.URL + paths.FullPath(paths.HealthPath)
	healthResponse := health.Health{}
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not complete request for %s: %w", paths.HealthPath, err)
	}
	defer resp.Body.Close()

	status := resp.StatusCode
	if status != http.StatusOK {
		if err, ok := iamerrors.Codes[status]; ok {
			return nil, err
		}
		return nil, fmt.Errorf("unhandled error returned http %d: %w", status, iamerrors.ErrInternal)
	}

	err = json.NewDecoder(resp.Body).Decode(&healthResponse)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", paths.HealthPath, err)
	}
	return &healthResponse, nil
}

// Ready calls the ready endpoint.
func (c *Client) Ready() error {
	url := c.URL + paths.FullPath(paths.ReadyPath)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return fmt.Errorf("could not complete request for %s: %w", paths.ReadyPath, err)
	}
	defer resp.Body.Close()

	status := resp.StatusCode
	if status != http.StatusOK {
		if err, ok := iamerrors.Codes[status]; ok {
			return err
		}
		return fmt.Errorf("unhandled error returned http %d: %w", status, iamerrors.ErrInternal)
	}
	return nil
}
