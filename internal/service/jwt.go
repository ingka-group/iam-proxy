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
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"

	"github.com/ingka-group-digital/iam-proxy/internal/models"
)

const (
	issuer             = "iam-proxy"
	invalidTokenError  = "token is invalid"
	invalidIssuer      = "issuer is invalid"
	parseTokenError    = "could not parse token"
	expirationInterval = 1 * time.Hour
)

// Claims defines the token claims.
type Claims struct {
	jwt.StandardClaims
}

// verifyUser checks the iam privileges for the given client id and secret.
func (s *Service) verifyUser(clientID, clientSecret string) (string, error) {
	if len(clientID) == 0 {
		return "", fmt.Errorf("no client Id provided")
	}
	if len(clientSecret) == 0 {
		return "", fmt.Errorf("no client secret provided")
	}

	clID := models.ClientID(clientID)
	if _, ok := s.IAM[clID]; !ok {
		return "", fmt.Errorf("client id does not exist")
	}
	secret := s.IAM[clID].ClientSecret
	if secret != clientSecret {
		return "", fmt.Errorf("client secret does not match")
	}
	return s.IAM[clID].AppName, nil
}

// GenerateToken generates a secret token for the provided app.
func (s *Service) GenerateToken(_ context.Context, clientID, clientSecret string) (string, string, int64, error) {
	appName, err := s.verifyUser(clientID, clientSecret)
	if err != nil {
		return "", "", 0, fmt.Errorf("user not authorized to use iam service: %w", err)
	}

	expiration := time.Now().Add(expirationInterval).Unix()

	accessToken, err := s.createToken(&jwt.StandardClaims{
		Id:        uuid.New().String(),
		ExpiresAt: expiration,
		Issuer:    issuer,
	})
	if err != nil {
		return "", "", 0, fmt.Errorf("could not generate access token for %s: %w", appName, err)
	}

	identityToken, err := s.createToken(&jwt.StandardClaims{
		Issuer:  issuer,
		Subject: appName,
	})
	if err != nil {
		return "", "", 0, fmt.Errorf("could not generate identity token for %s: %w", appName, err)
	}

	return accessToken, identityToken, int64(expirationInterval.Seconds()), nil
}

func (s *Service) createToken(claims *jwt.StandardClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("could not generate token: %w", err)
	}
	return tokenString, nil
}

// ParseToken parses the token and confirms its validity.
func (s *Service) ParseToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return "", fmt.Errorf("%s: %w", parseTokenError, err)
	}

	if claims, ok := token.Claims.(*jwt.StandardClaims); ok && token.Valid {
		if claims.Issuer != issuer {
			return "", fmt.Errorf(invalidIssuer)
		}
		return claims.Subject, nil
	}

	return "", fmt.Errorf(invalidTokenError)
}
