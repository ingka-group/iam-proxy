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

// Code generated by gowrap. DO NOT EDIT.
// template: ../../templates/opencensus_metrics.tpl
// gowrap: http://github.com/hexdigest/gowrap

import (
	"context"
	"log"
	"time"

	"github.com/ingka-group-digital/iam-proxy/client/health"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	servicerHistogramResultTag       = tag.MustNewKey("result")
	servicerHistogramMethodNameTag   = tag.MustNewKey("method_name")
	servicerHistogramInstanceNameTag = tag.MustNewKey("instance_name")

	servicerHistogram = stats.Float64("servicer_servicer_duration_ms", "Time spent in ms on running an operation of type <Servicer> in package <servicer>", stats.UnitMilliseconds)
)

// ServicerWithMetrics implements Servicer interface with all methods wrapped
// with metrics
type ServicerWithMetrics struct { //nolint:golint
	base         Servicer
	instanceName string
}

// NewWithMetrics returns an instance of the Servicer decorated with metrics
func NewWithMetrics(base Servicer, instanceName string) Servicer {
	if err := view.Register(
		&view.View{
			Name:        servicerHistogram.Name(),
			Description: servicerHistogram.Description(),
			Measure:     servicerHistogram,
			TagKeys: []tag.Key{
				servicerHistogramResultTag,
				servicerHistogramMethodNameTag,
				servicerHistogramInstanceNameTag,
			},
			Aggregation: ochttp.DefaultLatencyDistribution,
		},
	); err != nil {
		log.Fatalf("could not register view (%v): %v", servicerHistogram.Name(), err)
	}

	return ServicerWithMetrics{
		base:         base,
		instanceName: instanceName,
	}
}

// GenerateToken implements Servicer
func (_d ServicerWithMetrics) GenerateToken(ctx context.Context, key string, secret string) (s1 string, s2 string, i1 int64, err error) {
	_since := time.Now()
	defer func() {
		result := "ok"
		if err != nil {
			result = "error"
		}

		_ctx, err := tag.New(context.Background(),
			tag.Insert(servicerHistogramInstanceNameTag, _d.instanceName),
			tag.Insert(servicerHistogramMethodNameTag, "GenerateToken"),
			tag.Insert(servicerHistogramResultTag, result),
		)
		if err != nil {
			log.Printf("could not create tag with context for instance (%v) method (%v): %v",
				_d.instanceName,
				"GenerateToken",
				err,
			)
			return
		}
		stats.Record(
			_ctx,
			servicerHistogram.M(float64(time.Since(_since)/time.Millisecond)),
		)
	}()

	return _d.base.GenerateToken(ctx, key, secret)
}

// Health implements Servicer
func (_d ServicerWithMetrics) Health(ctx context.Context) (h1 health.Health, err error) {
	_since := time.Now()
	defer func() {
		result := "ok"
		if err != nil {
			result = "error"
		}

		_ctx, err := tag.New(context.Background(),
			tag.Insert(servicerHistogramInstanceNameTag, _d.instanceName),
			tag.Insert(servicerHistogramMethodNameTag, "Health"),
			tag.Insert(servicerHistogramResultTag, result),
		)
		if err != nil {
			log.Printf("could not create tag with context for instance (%v) method (%v): %v",
				_d.instanceName,
				"Health",
				err,
			)
			return
		}
		stats.Record(
			_ctx,
			servicerHistogram.M(float64(time.Since(_since)/time.Millisecond)),
		)
	}()

	return _d.base.Health(ctx)
}

// ParseToken implements Servicer
func (_d ServicerWithMetrics) ParseToken(tokenString string) (s1 string, err error) {
	_since := time.Now()
	defer func() {
		result := "ok"
		if err != nil {
			result = "error"
		}

		_ctx, err := tag.New(context.Background(),
			tag.Insert(servicerHistogramInstanceNameTag, _d.instanceName),
			tag.Insert(servicerHistogramMethodNameTag, "ParseToken"),
			tag.Insert(servicerHistogramResultTag, result),
		)
		if err != nil {
			log.Printf("could not create tag with context for instance (%v) method (%v): %v",
				_d.instanceName,
				"ParseToken",
				err,
			)
			return
		}
		stats.Record(
			_ctx,
			servicerHistogram.M(float64(time.Since(_since)/time.Millisecond)),
		)
	}()

	return _d.base.ParseToken(tokenString)
}

// Ready implements Servicer
func (_d ServicerWithMetrics) Ready(ctx context.Context) (err error) {
	_since := time.Now()
	defer func() {
		result := "ok"
		if err != nil {
			result = "error"
		}

		_ctx, err := tag.New(context.Background(),
			tag.Insert(servicerHistogramInstanceNameTag, _d.instanceName),
			tag.Insert(servicerHistogramMethodNameTag, "Ready"),
			tag.Insert(servicerHistogramResultTag, result),
		)
		if err != nil {
			log.Printf("could not create tag with context for instance (%v) method (%v): %v",
				_d.instanceName,
				"Ready",
				err,
			)
			return
		}
		stats.Record(
			_ctx,
			servicerHistogram.M(float64(time.Since(_since)/time.Millisecond)),
		)
	}()

	return _d.base.Ready(ctx)
}
