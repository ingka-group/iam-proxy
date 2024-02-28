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

package models

import "time"

// IAM is the collection of all valid client credentials and tokens.
type IAM map[ClientID]Secret

// ClientID defines a client id token.
type ClientID string

// Secret holds information on the client secret key.
type Secret struct {
	// ClientSecret is the client secret key for the corresponding client id
	ClientSecret string `json:"client_secret"`
	// AppName defines the application that uses this credential
	AppName        string    `json:"app_name"`
	ExpirationDate time.Time `json:"-"`
}
