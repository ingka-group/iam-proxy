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

package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ingka-group-digital/iam-proxy/client/health"
	"github.com/ingka-group-digital/iam-proxy/client/jwt"
	"github.com/ingka-group-digital/iam-proxy/internal/config"
	"github.com/ingka-group-digital/iam-proxy/internal/models"
	"github.com/ingka-group-digital/iam-proxy/internal/testutil"
)

func TestIAMConfigurationInit(t *testing.T) {

	type test struct {
		iam config.IAM
		err bool
		len int
	}

	tests := map[string]test{
		"init empty": {
			iam: config.IAM{
				Users: "",
			},
			err: true,
		},
		"init bad json": {
			iam: config.IAM{
				Users: "{,}",
			},
			err: true,
		},
		"init ok": {
			iam: config.IAM{
				Users: jwt.Base64Encode([]byte(`{"<client_id>" : { "client_secret" : "<client_secret>" , "app_name" : "<demo>" } }`)),
			},
		},
		"init ok with many": {
			iam: config.IAM{
				Users: jwt.Base64Encode([]byte(`{"<client_id>" : { "client_secret" : "<client_secret>" , "app_name" : "<demo>" } , "<client_id-2>" : { "client_secret" : "<client_secret-2>" , "app_name" : "<demo-2>" } }`)),
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			srv, err := New(defaultConfig(tt.iam))
			if tt.err {
				assert.Error(t, err)
				assert.Nil(t, srv)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, srv)

			if service, ok := srv.(*Service); ok {
				if tt.len > 0 {
					assert.Equal(t, tt.len, len(service.IAM))
				}
			} else {
				assert.Fail(t, "found wrong implementation of interface Servicer")
			}
		})
	}
}

func TestAuth2Token(t *testing.T) {

	type test struct {
		iam          config.IAM
		clientID     string
		clientSecret string
		err          bool
	}

	tests := map[string]test{
		"ok": {
			iam: config.IAM{
				Users: jwt.Base64Encode([]byte(`{"<client_id>" : { "client_secret" : "<client_secret>" , "app_name" : "<demo>" } }`)),
			},
			clientID:     "<client_id>",
			clientSecret: "<client_secret>",
		},
		"wrong user id": {
			iam: config.IAM{
				Users: jwt.Base64Encode([]byte(`{"<client_id>" : { "client_secret" : "<client_secret>" , "app_name" : "<demo>" } }`)),
			},
			err:          true,
			clientID:     "<other_client_id>",
			clientSecret: "<client_secret>",
		},
		"wrong user secret": {
			iam: config.IAM{
				Users: jwt.Base64Encode([]byte(`{"<client_id>" : { "client_secret" : "<client_secret>" , "app_name" : "<demo>" } }`)),
			},
			err:          true,
			clientID:     "<client_id>",
			clientSecret: "<other_client_secret>",
		},
		"ok with sha secret": {
			iam: config.IAM{
				Users:  jwt.Base64Encode([]byte(`{"<client_id>" : { "client_secret" : "<client_secret>" , "app_name" : "<demo>" } }`)),
				Secret: jwt.CryptoHash("test"),
			},
			clientID:     "<client_id>",
			clientSecret: "<client_secret>",
		},
		"ok with string secret": {
			iam: config.IAM{
				Users:  jwt.Base64Encode([]byte(`{"<client_id>" : { "client_secret" : "<client_secret>" , "app_name" : "<demo>" } }`)),
				Secret: "test",
			},
			clientID:     "<client_id>",
			clientSecret: "<client_secret>",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			srv, err := New(defaultConfig(tt.iam))

			assert.NoError(t, err)
			assert.NotNil(t, srv)

			token, _, expiration, err := srv.GenerateToken(context.TODO(), tt.clientID, tt.clientSecret)

			if tt.err {
				assert.Error(t, err)
				assert.Empty(t, token)
				assert.Equal(t, int64(0), expiration)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, token)
			assert.Equal(t, int64(expirationInterval.Seconds()), expiration)

		})
	}
}

func TestService_Health(t *testing.T) {

	type test struct {
		iam    config.IAM
		status health.Status
	}

	tests := map[string]test{
		"alive": {
			iam: config.IAM{
				Users:  jwt.Base64Encode([]byte(`{"<client_id>" : { "client_secret" : "<client_secret>" , "app_name" : "<demo>" } }`)),
				Secret: "test",
			},
			status: health.StatusAlive,
		},
		"degraded": {
			iam: config.IAM{
				Users: jwt.Base64Encode([]byte(`{"<client_id>" : { "client_secret" : "<client_secret>" , "app_name" : "<demo>" } }`)),
			},
			status: health.StatusDegraded,
		},
		"unavailable": {
			iam: config.IAM{
				Users:  jwt.Base64Encode([]byte(`{}`)),
				Secret: "test",
			},
			status: health.StatusUnavailable,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			srv, err := New(defaultConfig(tt.iam))

			assert.NoError(t, err)
			assert.NotNil(t, srv)

			h, err := srv.Health(context.TODO())

			assert.NoError(t, err)
			assert.Equal(t, tt.status, h.Status)

		})
	}

}

func TestCredentialsStruct(t *testing.T) {

	sToken := "demo-token"
	secretToken := jwt.CryptoHash(sToken)
	assert.NotEmpty(t, secretToken)

	cid := "demo-client"
	clientID := jwt.CryptoHash(cid)

	cSecret := "demo-secret"
	clientSecret := jwt.CryptoHash(cSecret)

	iam := models.IAM{}
	iam[models.ClientID(clientID)] = models.Secret{
		AppName:      "demo-app",
		ClientSecret: clientSecret,
	}

	b, err := json.Marshal(iam)
	assert.NoError(t, err)

	credentials := `{"dggA6Hiw32JkOfXRmnNo3vynK3M":{"client_secret":"RRxJIyRVQ6YF4SraV4gSdmA9s1s","app_name":"demo-app"}}`
	// Note : with a non=raw encoding strategy for the cryptohash function we would have an '=' sign at the end of each hash
	assert.Equal(t, credentials, string(b))

	base64 := jwt.Base64Encode([]byte(credentials))
	decredentials, err := jwt.Base64Decode(base64)
	assert.NoError(t, err)
	assert.Equal(t, string(decredentials), credentials)
}

func defaultConfig(iam config.IAM) Config {
	cfg := testutil.SampleConfig()
	cfg.IAM = iam
	return Config{
		Config: cfg,
	}
}
