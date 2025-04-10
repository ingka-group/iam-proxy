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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	clienthttp "github.com/ingka-group/iam-proxy/client/http"
	"github.com/ingka-group/iam-proxy/client/iam"
	"github.com/ingka-group/iam-proxy/client/jwt"
	"github.com/ingka-group/iam-proxy/client/paths"
	"github.com/ingka-group/iam-proxy/internal/config"
	"github.com/ingka-group/iam-proxy/internal/service"
	"github.com/ingka-group/iam-proxy/internal/testutil"
)

func TestClient_GenerateToken(t *testing.T) {
	t.Parallel()
	type args struct {
		cfg Config
	}

	tests := []struct {
		name       string
		args       args
		wantCode   int
		wantErr    bool
		parsingErr bool
		iam        config.IAM
		body       string
	}{
		{
			name: "token",
			args: args{
				cfg: Config{
					Config: testutil.SampleConfig(),
				},
			},
			iam: config.IAM{
				Users: jwt.Base64Encode([]byte(`{"<client_id>" : { "client_secret" : "<client_secret>" , "app_name" : "<ocp>" } }`)),
			},
			body:     "client_id=<client_id>&client_secret=<client_secret>&grant_type=client_credentials",
			wantCode: 200,
		},
		{
			name: "client_id_missing",
			args: args{
				cfg: Config{
					Config: testutil.SampleConfig(),
				},
			},
			iam: config.IAM{
				Users: jwt.Base64Encode([]byte(`{"<client_id>" : { "client_secret" : "<client_secret>" , "app_name" : "<ocp>" } }`)),
			},
			body:       "client_secret=<client_secret>&grant_type=client_credentials",
			parsingErr: true,
			wantCode:   400,
		},
		{
			name: "client_secret_missing",
			args: args{
				cfg: Config{
					Config: testutil.SampleConfig(),
				},
			},
			iam: config.IAM{
				Users: jwt.Base64Encode([]byte(`{"<client_id>" : { "client_secret" : "<client_secret>" , "app_name" : "<ocp>" } }`)),
			},
			body:       "client_id=<client_id>",
			parsingErr: true,
			wantCode:   400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, err := service.New(defaultConfig(tt.iam))
			if err != nil {
				t.Errorf("could not create service instance: %s", err.Error())
			}

			tt.args.cfg.Service = srv
			c, err := New(tt.args.cfg)
			if err != nil {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
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

func TestClient_ParseToken(t *testing.T) {
	t.Parallel()
	type args struct {
		cfg Config
	}

	tests := []struct {
		name     string
		args     args
		wantCode int
		wantErr  bool
		iam      config.IAM
		header   map[string]string
	}{
		{
			name: "token empty",
			args: args{
				cfg: Config{
					Config: testutil.SampleConfig(),
				},
			},
			iam: config.IAM{
				Users: jwt.Base64Encode([]byte(`{"<client_id>" : { "client_secret" : "<client_secret>" , "app_name" : "<ocp>" } }`)),
			},
			header:   map[string]string{},
			wantCode: 400,
		},
		{
			name: "wrong header",
			args: args{
				cfg: Config{
					Config: testutil.SampleConfig(),
				},
			},
			iam: config.IAM{
				Users: jwt.Base64Encode([]byte(`{"<client_id>" : { "client_secret" : "<client_secret>" , "app_name" : "<ocp>" } }`)),
			},
			header: map[string]string{
				"Id": "token",
			},
			wantCode: 400,
		},
		{
			name: "no auth word",
			args: args{
				cfg: Config{
					Config: testutil.SampleConfig(),
				},
			},
			iam: config.IAM{
				Users: jwt.Base64Encode([]byte(`{"<client_id>" : { "client_secret" : "<client_secret>" , "app_name" : "<ocp>" } }`)),
			},
			header: map[string]string{
				clienthttp.AuthorizationHeaderKey: "token",
			},
			wantCode: 400,
		},
		{
			name: "random token",
			args: args{
				cfg: Config{
					Config: testutil.SampleConfig(),
				},
			},
			iam: config.IAM{
				Users: jwt.Base64Encode([]byte(`{"<client_id>" : { "client_secret" : "<client_secret>" , "app_name" : "<ocp>" } }`)),
			},
			header: map[string]string{
				clienthttp.AuthorizationHeaderKey: "Authorization token",
			},
			wantCode: 401,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, err := service.New(defaultConfig(tt.iam))
			if err != nil {
				t.Errorf("could not create service instance: %s", err.Error())
			}

			tt.args.cfg.Service = srv
			c, err := New(tt.args.cfg)
			if err != nil {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
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

func TestClient_IdentityToken(t *testing.T) {
	t.Parallel()
	type args struct {
		cfg Config
	}

	tests := []struct {
		name     string
		args     args
		wantCode int
		wantErr  bool
		iam      config.IAM
		header   map[string]string
		identity string
	}{
		{
			name: "token empty",
			args: args{
				cfg: Config{
					Config: testutil.SampleConfig(),
				},
			},
			iam: config.IAM{
				Users: jwt.Base64Encode([]byte(`{"<client_id>" : { "client_secret" : "<client_secret>" , "app_name" : "<ocp>" } }`)),
			},
			header:   map[string]string{},
			wantCode: 400,
		},
		{
			name: "wrong header",
			args: args{
				cfg: Config{
					Config: testutil.SampleConfig(),
				},
			},
			iam: config.IAM{
				Users: jwt.Base64Encode([]byte(`{"<client_id>" : { "client_secret" : "<client_secret>" , "app_name" : "<ocp>" } }`)),
			},
			header: map[string]string{
				"Id": "token",
			},
			wantCode: 400,
		},
		{
			name: "no auth word",
			args: args{
				cfg: Config{
					Config: testutil.SampleConfig(),
				},
			},
			iam: config.IAM{
				Users: jwt.Base64Encode([]byte(`{"<client_id>" : { "client_secret" : "<client_secret>" , "app_name" : "<ocp>" } }`)),
			},
			header: map[string]string{
				clienthttp.IdentityHeaderKey: "token",
			},
			wantCode: 400,
		},
		{
			name: "random token",
			args: args{
				cfg: Config{
					Config: testutil.SampleConfig(),
				},
			},
			iam: config.IAM{
				Users: jwt.Base64Encode([]byte(`{"<client_id>" : { "client_secret" : "<client_secret>" , "app_name" : "<ocp>" } }`)),
			},
			header: map[string]string{
				clienthttp.IdentityHeaderKey: "Identity token",
			},
			wantCode: 401,
		},
		{
			name: "correct token",
			args: args{
				cfg: Config{
					Config: testutil.SampleConfig(),
				},
			},
			iam: config.IAM{
				Users: jwt.Base64Encode([]byte(`{"<client_id>" : { "client_secret" : "<client_secret>" , "app_name" : "<ocp>" } }`)),
			},
			wantCode: 200,
			identity: "<ocp>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, err := service.New(defaultConfig(tt.iam))
			if err != nil {
				t.Errorf("could not create service instance: %s", err.Error())
			}

			tt.args.cfg.Service = srv
			c, err := New(tt.args.cfg)
			if err != nil {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.header == nil {
				// generate an identity token
				_, id, _, err := srv.GenerateToken(context.TODO(), "<client_id>", "<client_secret>")
				assert.NoError(t, err)
				tt.header = map[string]string{
					clienthttp.IdentityHeaderKey: fmt.Sprintf("%s %s", clienthttp.IdentityHeaderKey, id),
				}
			}

			resp, err := doRequest("POST", paths.FullPath(paths.Identity), nil, tt.header, c)
			if err != nil {
				t.Error("Failed to perform request", err)
			}

			if resp.Code != tt.wantCode {
				t.Errorf("Expected return code %v but got %v", tt.wantCode, resp.Code)
			}

			if resp.Code == 200 {
				m := make(map[string]string)
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				// parse as token response
				err = json.Unmarshal(bodyBytes, &m)
				assert.NoError(t, err)
				assert.Equal(t, m["identity"], tt.identity)
			}

		})
	}
}

func TestClient_End2End(t *testing.T) {
	cfg := config.IAM{
		Users: jwt.Base64Encode([]byte(`{"<client_id>" : { "client_secret" : "<client_secret>" , "app_name" : "<ocp>" } }`)),
	}

	srv, err := service.New(defaultConfig(cfg))
	if err != nil {
		t.Errorf("could not create service instance: %s", err.Error())
	}

	c, err := New(Config{
		Config:  testutil.SampleConfig(),
		Service: srv,
	})
	assert.NoError(t, err)

	// generate the token
	resp, err := doRequest("POST", paths.FullPath(paths.OAuthToken), []byte("client_id=<client_id>&client_secret=<client_secret>&grant_type=client_credentials"), map[string]string{}, c)
	if err != nil {
		t.Error("Failed to perform `generate token` request", err)
	}

	token := new(iam.Token)

	if resp.Code == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("failed to parse response body")
		}
		// parse as token response
		err = json.Unmarshal(bodyBytes, token)
		assert.NoError(t, err)
	}

	resp, err = doRequest("POST", paths.FullPath(paths.ValidateToken), nil, map[string]string{
		clienthttp.AuthorizationHeaderKey: fmt.Sprintf("Authorization %s", token.AccessToken),
	}, c)
	if err != nil {
		t.Error("Failed to perform `validate token` request", err)
	}

	assert.Equal(t, resp.Code, http.StatusOK)
}

func defaultConfig(iam config.IAM) service.Config {
	cfg := testutil.SampleConfig()
	cfg.IAM = iam
	return service.Config{
		Config: cfg,
	}
}
