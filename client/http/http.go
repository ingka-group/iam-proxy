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

package jwt

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	// AuthorizationHeaderKey is the authorization header key
	AuthorizationHeaderKey = "Authorization"
	// IdentityHeaderKey is the identity header key
	IdentityHeaderKey = "Identity"
)

// InsertAccessToken inserts the access token correctly formatted into the request header
func InsertAccessToken(r *http.Request, token string) {
	insertToken(r, AuthorizationHeaderKey, token)
}

// InsertIdentityToken inserts the access token correctly formatted into the request header
func InsertIdentityToken(r *http.Request, token string) {
	insertToken(r, IdentityHeaderKey, token)
}

func insertToken(r *http.Request, key, token string) {
	if r.Header == nil {
		r.Header = make(map[string][]string)
	}
	//normally : `Authorization the_token_xxx`
	r.Header.Set(key, fmt.Sprintf("%s %s", key, token))
}

// ExtractAccessToken extracts the access token from the request header.
func ExtractAccessToken(r *http.Request) (string, error) {
	return extractToken(r, AuthorizationHeaderKey)
}

// ExtractIdentityToken extracts the identity token from the request header.
func ExtractIdentityToken(r *http.Request) (string, error) {
	return extractToken(r, IdentityHeaderKey)
}

func extractToken(r *http.Request, key string) (string, error) {
	bearToken := r.Header.Get(key)
	//normally : `Authorization the_token_xxx`
	if len(bearToken) == 0 {
		return "", fmt.Errorf("token not present in header")
	}
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1], nil
	}
	return "", fmt.Errorf("bad header format")
}
