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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestHeader(t *testing.T) {
	req := new(http.Request)
	insertToken(req, "key", "token")

	token, err := extractToken(req, "key")
	assert.NoError(t, err)
	assert.Equal(t, "token", token)
}
