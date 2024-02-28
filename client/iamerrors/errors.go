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

package iamerrors

import (
	"errors"
	"net/http"
)

var (
	// ErrServiceUnavailable is the error signifying when the service is not available.
	ErrServiceUnavailable = errors.New("service is unavailable")
	// ErrInternal defines a generic internal server error.
	ErrInternal = errors.New("internal server error")
	// ErrUnauthorized defines an error that signifies that the client is not authorised to use the service.
	ErrUnauthorized = errors.New("user credentials invalid")
	// ErrBadRequest defines an error that signifies that the client request is not valid.
	ErrBadRequest = errors.New("bad request")
	// ErrBadResponse defines an error that signifies that the response could not be properly parsed.
	ErrBadResponse = errors.New("bad response")
)

// Codes are the service error codes mapping.
var Codes = map[int]error{
	http.StatusServiceUnavailable:  ErrServiceUnavailable,
	http.StatusInternalServerError: ErrInternal,
	http.StatusBadRequest:          ErrBadRequest,
	http.StatusUnauthorized:        ErrUnauthorized,
}
