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
	"encoding/json"
	"fmt"

	"github.com/ingka-group-digital/iam-proxy/client/jwt"
	"github.com/ingka-group-digital/iam-proxy/internal/config"
	"github.com/ingka-group-digital/iam-proxy/internal/models"
)

// Config for iam-proxy-v1 Service
type Config struct {
	*config.Config
}

// New creates and initializes iam-proxy-v1 Service
func New(c Config) (Servicer, error) {
	iam := new(models.IAM)
	// load IAM configuration
	iamUsers, err := jwt.Base64Decode(c.IAM.Users)
	if err != nil {
		return nil, fmt.Errorf("could not decode iam users credentials: %w", err)
	}
	err = json.Unmarshal(iamUsers, iam)
	if err != nil {
		return nil, fmt.Errorf("could not decode IAM information: %w", err)
	}

	for _, u := range *iam {
		c.Logger.Infof("loaded user credentials for %s", u.AppName)
	}

	return &Service{
		IAM:    *iam,
		secret: []byte(c.IAM.Secret),
	}, nil
}
