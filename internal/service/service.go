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

//go:generate mockgen --build_flags=--mod=mod -destination=./mock_service/mock_service.go -source=service.go -copyright_file=../../templates/copyright-header.tpl
//go:generate gowrap gen -g -i Servicer -t ../../templates/opentelemetry.tpl -o servicer_with_tracing_gen.go
//go:generate gowrap gen -g -i Servicer -t ../../templates/opencensus_metrics.tpl -o servicer_with_metrics_gen.go

import (
	"context"

	"github.com/ingka-group/iam-proxy/client/health"
	"github.com/ingka-group/iam-proxy/internal/models"
)

// Servicer interface of service
type Servicer interface {
	Health(ctx context.Context) (health.Health, error)
	Ready(ctx context.Context) error
	GenerateToken(ctx context.Context, key, secret string) (string, string, int64, error)
	ParseToken(tokenString string) (string, error)
}

// Service implements business logic of iam-proxy-v1 Service
type Service struct {
	Config
	IAM    models.IAM
	secret []byte
}

// Health performs health checks and returns the health of the service
func (s *Service) Health(_ context.Context) (health.Health, error) {
	if len(s.IAM) == 0 {
		return health.Health{
			Status: health.StatusUnavailable,
			IAM:    "No client credentials available",
		}, nil
	}
	if len(s.secret) == 0 {
		return health.Health{
			Status: health.StatusDegraded,
			IAM:    "insecure secret",
		}, nil
	}
	return health.Health{
		Status: health.StatusAlive,
	}, nil
}

// Ready returns non-nil response if service is ready to receive requests
func (s *Service) Ready(_ context.Context) error {
	return nil
}
