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

package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"

	"github.com/ingka-group/iam-proxy/client/health"
	"github.com/ingka-group/iam-proxy/client/paths"
	"github.com/ingka-group/iam-proxy/internal/config"
	"github.com/ingka-group/iam-proxy/internal/logger"
	"github.com/ingka-group/iam-proxy/internal/service/mock_service"
	"github.com/ingka-group/iam-proxy/internal/testutil"
)

func TestMain(m *testing.M) {
	logger.Logger = zap.NewExample()
	os.Exit(m.Run())
}

func TestHealth(t *testing.T) {
	t.Parallel()

	type args struct {
		cfg  Config
		mock *mock_service.MockServicer
	}
	ctrl := gomock.NewController(t)
	tests := []struct {
		name     string
		args     args
		want     health.Health
		wantCode int
		wantErr  bool
	}{
		{
			name: "health_" + config.ServiceName,
			args: args{
				cfg: Config{
					Config: testutil.SampleConfig(),
				},
				mock: mock_service.NewMockServicer(ctrl),
			},
			wantCode: 200,
			want:     health.Health{Status: health.StatusAlive},
		},
		{
			name: "unhealthy_" + config.ServiceName,
			args: args{
				cfg: Config{
					Config: testutil.SampleConfig(),
				},
				mock: mock_service.NewMockServicer(ctrl),
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name: "degraded_" + config.ServiceName,
			args: args{
				cfg: Config{
					Config: testutil.SampleConfig(),
				},
				mock: mock_service.NewMockServicer(ctrl),
			},
			wantCode: http.StatusServiceUnavailable,
			want:     health.Health{Status: health.StatusUnavailable},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cfg.Service = tt.args.mock
			c, err := New(tt.args.cfg)
			if err != nil {
				t.Errorf("unexpected error on create server")
			}

			if tt.wantErr {
				tt.args.mock.EXPECT().Health(gomock.Any()).Return(tt.want, errors.New("some error"))
			} else {
				tt.args.mock.EXPECT().Health(gomock.Any()).Return(tt.want, nil)
			}

			resp, err := doRequest("GET", paths.FullPath(paths.HealthPath), nil, map[string]string{}, c)
			if err != nil {
				t.Error("Failed to perform request", err)
			}

			if resp.Code != tt.wantCode {
				t.Errorf("Expected return code %v but got %v", tt.wantCode, resp.Code)
			}

			var h health.Health
			if err := json.NewDecoder(resp.Body).Decode(&h); (err != nil) != tt.wantErr {
				t.Error("Cannot decode response", err)
			}

			if diff := cmp.Diff(h, tt.want); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func TestReady(t *testing.T) {
	t.Parallel()
	type args struct {
		cfg  Config
		mock *mock_service.MockServicer
	}
	ctrl := gomock.NewController(t)
	tests := []struct {
		name     string
		args     args
		wantCode int
		wantErr  bool
	}{
		{
			name: "ready",
			args: args{
				cfg: Config{
					Config: testutil.SampleConfig(),
				},
				mock: mock_service.NewMockServicer(ctrl),
			},
			wantCode: 200,
		},
		{
			name: "not-ready",
			args: args{
				cfg: Config{
					Config: testutil.SampleConfig(),
				},
				mock: mock_service.NewMockServicer(ctrl),
			},
			wantErr:  true,
			wantCode: http.StatusServiceUnavailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cfg.Service = tt.args.mock
			c, err := New(tt.args.cfg)
			if err != nil {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				tt.args.mock.EXPECT().Ready(gomock.Any()).Return(errors.New("some error"))
			} else {
				tt.args.mock.EXPECT().Ready(gomock.Any()).Return(nil)
			}

			resp, err := doRequest("GET", paths.FullPath(paths.ReadyPath), nil, map[string]string{}, c)
			if err != nil {
				t.Error("Failed to perform request", err)
			}

			if resp.Code != tt.wantCode {
				t.Errorf("Expected return code %v but got %v", tt.wantCode, resp.Code)
			}
		})
	}
}

func doRequest(method, uri string, body []byte, header map[string]string, c *Client) (*httptest.ResponseRecorder, error) {
	resp := httptest.NewRecorder()

	req, err := http.NewRequest(method, uri, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	requestHeader := make(http.Header)
	for n, h := range header {
		requestHeader[n] = []string{h}
	}
	req.Header = requestHeader

	router := c.setupRouter()
	router.ServeHTTP(resp, req)

	return resp, nil
}
