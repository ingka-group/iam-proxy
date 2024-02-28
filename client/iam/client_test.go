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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ingka-group-digital/iam-proxy/client/health"
	"github.com/ingka-group-digital/iam-proxy/client/paths"
)

func TestClient_Health(t *testing.T) {
	t.Parallel()

	type test struct {
		response   health.Health
		statusCode int
		err        bool
	}

	tests := map[string]test{
		"response": {
			response: health.Health{
				Status: "Active",
				IAM:    "my:iam",
			},
			statusCode: http.StatusOK,
		},
		"error": {
			statusCode: http.StatusServiceUnavailable,
			err:        true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			url, port, clb := mockServerWithResponse(t, mustEncode(tt.response), tt.statusCode, assertURL(paths.FullPath(paths.HealthPath)))
			defer clb()

			client := New(fmt.Sprintf("http://%s:%d", url, port), http.DefaultClient)

			health, err := client.Health()
			if tt.err {
				assert.Error(t, err)
				assert.Nil(t, health)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.response, *health)
			}
		})
	}
}

func TestClient_Ready(t *testing.T) {
	t.Parallel()

	type test struct {
		statusCode int
		err        bool
	}

	tests := map[string]test{
		"response": {
			statusCode: http.StatusOK,
		},
		"error": {
			statusCode: http.StatusServiceUnavailable,
			err:        true,
		},
	}

	for name, tt := range tests {
		// pin loop variable
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			url, port, clb := mockServerWithResponse(t, "", tt.statusCode, assertURL(paths.FullPath(paths.ReadyPath)))
			defer clb()

			client := New(fmt.Sprintf("http://%s:%d", url, port), http.DefaultClient)

			err := client.Ready()
			if tt.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func mockServerWithResponse(t *testing.T, response string, code int, requestAssertions ...func(t *testing.T, request *http.Request)) (string, int, func()) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		for _, assertion := range requestAssertions {
			assertion(t, request)
		}
		writer.WriteHeader(code)
		fmt.Fprint(writer, response)
	}))

	serviceURL, err := url.Parse(ts.URL)

	assert.NoError(t, err)

	port, err := strconv.Atoi(serviceURL.Port())
	assert.NoError(t, err)

	return serviceURL.Hostname(), port, ts.Close
}

func mustEncode(health health.Health) string {
	b, err := json.Marshal(health)
	if err != nil {
		log.Fatalf("could not encode test respose: %v", err)
	}
	return string(b)
}

func assertURL(url string) func(t *testing.T, request *http.Request) {
	return func(t *testing.T, request *http.Request) {
		assert.Equal(t, url, request.URL.Path)
	}
}
