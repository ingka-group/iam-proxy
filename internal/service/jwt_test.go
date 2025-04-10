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
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	jwtmodule "github.com/ingka-group/iam-proxy/client/jwt"
	"github.com/ingka-group/iam-proxy/internal/models"
)

const (
	testClientID1     = "<client_id_1>"
	testClientSecret1 = "<client_secret_1>"
	testClientID2     = "<client_id_2>"
	testClientSecret2 = "<client_secret_2>"
)

func TestService_GenerateToken(t *testing.T) {
	srv := newTestService()

	ctx := context.TODO()

	token, ocp, _, err := srv.GenerateToken(ctx, testClientID1, testClientSecret1)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assertIdentity(t, "ocp", ocp)

	otherToken, atp, _, err := srv.GenerateToken(ctx, testClientID2, testClientSecret2)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assertIdentity(t, "atp", atp)

	assert.NotEqual(t, token, otherToken)
}

func TestService_ParseToken(t *testing.T) {

	srv := newTestService()

	token, _, _, err := srv.GenerateToken(context.TODO(), testClientID1, testClientSecret1)
	assert.NoError(t, err)

	_, err = srv.ParseToken(token)
	assert.NoError(t, err)
}

func TestService_ParseToken_MixedInstance(t *testing.T) {

	genService := newTestService()
	parseService := newTestService()

	token, _, _, err := genService.GenerateToken(context.TODO(), testClientID1, testClientSecret1)
	assert.NoError(t, err)

	_, err = parseService.ParseToken(token)
	assert.NoError(t, err)
}

func TestService_ParseToken_ParseError(t *testing.T) {

	srv := newTestService()

	_, err := srv.ParseToken("")
	assert.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), parseTokenError))
}

func TestService_ParseToken_Expired(t *testing.T) {
	srv := newTestService()

	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Second)),
		Issuer:    issuer,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(make([]byte, 0))
	assert.NoError(t, err)

	_, err = srv.ParseToken(tokenString)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "token is expired"))
	assert.True(t, strings.Contains(err.Error(), parseTokenError))
}

func TestService_ParseToken_IssuerError(t *testing.T) {
	srv := newTestService()

	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		Issuer:    "other-issuer",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(make([]byte, 0))
	assert.NoError(t, err)

	_, err = srv.ParseToken(tokenString)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), invalidIssuer))
}

func newTestService() Service {
	return Service{
		Config: Config{},
		IAM: map[models.ClientID]models.Secret{
			testClientID1: {
				AppName:      "ocp",
				ClientSecret: testClientSecret1,
			},
			testClientID2: {
				AppName:      "atp",
				ClientSecret: testClientSecret2,
			},
		},
	}
}

func assertIdentity(t *testing.T, exp, actual string) {
	claims, err := jwtmodule.DecodeToken(actual)
	assert.NoError(t, err)
	assert.Equal(t, exp, claims.Subject)
}
