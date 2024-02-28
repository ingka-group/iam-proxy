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
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt"
)

// Base64Encode encodes a byte array to a base 64 string.
func Base64Encode(bb []byte) string {
	return base64.RawStdEncoding.EncodeToString(bb)
}

// Base64Decode decodes a base 64 string to a byte array.
func Base64Decode(s string) ([]byte, error) {
	return base64.RawStdEncoding.DecodeString(s)
}

// CryptoHash generates a cryptographically safe hash based on the given string parameter.
func CryptoHash(base string) string {
	hasher := sha1.New()
	hasher.Write([]byte(base))
	sha := base64.RawStdEncoding.EncodeToString(hasher.Sum(nil))
	return sha
}

// DecodeToken decodes the non-secret part of the jwt token and extracts the associated claims.
// Note: this method does not verify the tokens validity, for this the corresponding endpoint of the iam service should be used.
func DecodeToken(token string) (*jwt.StandardClaims, error) {
	s := strings.Split(token, ".")
	if len(s) < 2 {
		return nil, fmt.Errorf("token format is wrong")
	}
	b, err := base64.RawStdEncoding.DecodeString(s[1])
	if err != nil {
		return nil, fmt.Errorf("could not decode token part")
	}
	m := new(jwt.StandardClaims)
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal claims")
	}
	return m, nil
}
