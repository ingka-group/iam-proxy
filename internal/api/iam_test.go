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
	"errors"
	"testing"

	"github.com/golang/mock/gomock"

	clienthttp "github.com/ingka-group/iam-proxy/client/http"
	"github.com/ingka-group/iam-proxy/client/paths"
	"github.com/ingka-group/iam-proxy/internal/service/mock_service"
	"github.com/ingka-group/iam-proxy/internal/testutil"
)

func TestClient_Generate(t *testing.T) {
	t.Parallel()
	type args struct {
		cfg  Config
		mock *mock_service.MockServicer
	}
	ctrl := gomock.NewController(t)
	tests := []struct {
		name       string
		args       args
		wantCode   int
		wantErr    bool
		parsingErr bool
		body       string
	}{
		{
			name: "token",
			args: args{
				cfg: Config{
					Config: testutil.SampleConfig(),
				},
				mock: mock_service.NewMockServicer(ctrl),
			},
			body:     "client_id=<your-client-id>&client_secret=<your-client-secret>&grant_type=client_credentials",
			wantCode: 200,
		},
		{
			name: "token_error",
			args: args{
				cfg: Config{
					Config: testutil.SampleConfig(),
				},
				mock: mock_service.NewMockServicer(ctrl),
			},
			body:     "client_id=<your-client-id>&client_secret=<your-client-secret>&grant_type=client_credentials",
			wantErr:  true,
			wantCode: 401,
		},
		{
			name: "client_id_missing",
			args: args{
				cfg: Config{
					Config: testutil.SampleConfig(),
				},
				mock: mock_service.NewMockServicer(ctrl),
			},
			body:       "client_secret=<your-client-secret>&grant_type=client_credentials",
			parsingErr: true,
			wantCode:   400,
		},
		{
			name: "client_secret_missing",
			args: args{
				cfg: Config{
					Config: testutil.SampleConfig(),
				},
				mock: mock_service.NewMockServicer(ctrl),
			},
			body:       "client_id=<your-client-id>",
			parsingErr: true,
			wantCode:   400,
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

			if !tt.parsingErr {
				if tt.wantErr {
					tt.args.mock.EXPECT().GenerateToken(gomock.Any(), gomock.Any(), gomock.Any()).Return("", "", int64(0), errors.New("some error"))
				} else {
					tt.args.mock.EXPECT().GenerateToken(gomock.Any(), gomock.Eq("<your-client-id>"), gomock.Eq("<your-client-secret>")).Return("access-token", "identity-token", int64(1), nil)
				}
			}

			resp, err := doRequest("POST", paths.FullPath(paths.OAuthToken), []byte(tt.body), map[string]string{}, c)
			if err != nil {
				t.Error("Failed to perform request", err)
			}

			if resp.Code != tt.wantCode {
				t.Errorf("Expected return code %v but got %v", tt.wantCode, resp.Code)
			}
		})
	}
}

func TestClient_Parse(t *testing.T) {
	t.Parallel()
	type args struct {
		cfg  Config
		mock *mock_service.MockServicer
	}
	ctrl := gomock.NewController(t)
	tests := []struct {
		name       string
		args       args
		wantCode   int
		wantErr    bool
		parsingErr bool
		header     map[string]string
	}{
		{
			name: "token",
			args: args{
				cfg: Config{
					Config: testutil.SampleConfig(),
				},
				mock: mock_service.NewMockServicer(ctrl),
			},
			header: map[string]string{
				clienthttp.AuthorizationHeaderKey: "Authorization token",
			},
			wantCode: 200,
		},
		{
			name: "token_error",
			args: args{
				cfg: Config{
					Config: testutil.SampleConfig(),
				},
				mock: mock_service.NewMockServicer(ctrl),
			},
			header: map[string]string{
				clienthttp.AuthorizationHeaderKey: "Authorization token",
			},
			wantErr:  true,
			wantCode: 401,
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

			if !tt.parsingErr {
				if tt.wantErr {
					tt.args.mock.EXPECT().ParseToken(gomock.Any()).Return("", errors.New("some error"))
				} else {
					tt.args.mock.EXPECT().ParseToken(gomock.Any()).Return("<subject>", nil)
				}
			}

			resp, err := doRequest("POST", paths.FullPath(paths.ValidateToken), nil, tt.header, c)
			if err != nil {
				t.Error("Failed to perform request", err)
			}

			if resp.Code != tt.wantCode {
				t.Errorf("Expected return code %v but got %v", tt.wantCode, resp.Code)
			}
		})
	}
}
