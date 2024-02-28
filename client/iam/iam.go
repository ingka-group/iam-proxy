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

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	jwt "github.com/ingka-group-digital/iam-proxy/client/http"
	"github.com/ingka-group-digital/iam-proxy/client/iamerrors"
	"github.com/ingka-group-digital/iam-proxy/client/paths"
)

// Token calls the iam service and returns the access token based on the provided clientID and clientSecret.
func (c *Client) Token(clientID, clientSecret string) (string, error) {
	url := c.URL + paths.FullPath(paths.OAuthToken)
	resp, err := c.HTTPClient.Post(url, "text/plain", bytes.NewBuffer(buildOauthRequestBody(clientID, clientSecret)))
	if err != nil {
		return "", fmt.Errorf("could not complete request for %s: %w", paths.OAuthToken, err)
	}
	defer resp.Body.Close()

	status := resp.StatusCode
	if status != http.StatusOK {
		if err, ok := iamerrors.Codes[status]; ok {
			return "", err
		}
		return "", fmt.Errorf("unhandled error returned http %d: %w", status, iamerrors.ErrInternal)
	}

	token := new(Token)
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", iamerrors.ErrServiceUnavailable
	}
	// parse as token response
	err = json.Unmarshal(bodyBytes, token)
	if err != nil {
		return "", iamerrors.ErrBadResponse
	}
	return token.AccessToken, nil
}

// Validate calls the iam service and validates the given token.
func (c *Client) Validate(token string) error {
	url := c.URL + paths.FullPath(paths.ValidateToken)
	req, _ := http.NewRequest("POST", url, nil)
	jwt.InsertAccessToken(req, token)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not complete request for %s: %w", paths.ValidateToken, err)
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

// Identity calls the iam service and validates the given token while returning the identity information e.g. claims subject.
func (c *Client) Identity(token string) (string, error) {
	url := c.URL + paths.FullPath(paths.Identity)
	req, _ := http.NewRequest("POST", url, nil)
	jwt.InsertIdentityToken(req, token)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not complete request for %s: %w", paths.ValidateToken, err)
	}
	defer resp.Body.Close()

	status := resp.StatusCode
	if status != http.StatusOK {
		if err, ok := iamerrors.Codes[status]; ok {
			return "", err
		}
		return "", fmt.Errorf("unhandled error returned http %d: %w", status, iamerrors.ErrInternal)
	}

	m := make(map[string]string)
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", iamerrors.ErrServiceUnavailable
	}
	// parse as token response
	err = json.Unmarshal(bodyBytes, &m)
	if err != nil {
		return "", fmt.Errorf("could not decode response: %w", err)
	}
	return m["subject"], nil
}

func buildOauthRequestBody(clientID, clientSecret string) []byte {
	return []byte(fmt.Sprintf("client_id=%s&client_secret=%s&grant_type=client_credentials", clientID, clientSecret))
}
