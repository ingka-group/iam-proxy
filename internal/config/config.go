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

package config

import (
	"context"
	"fmt"
	"time"

	"contrib.go.opencensus.io/exporter/ocagent"
	"github.com/blendle/zapdriver"
	"github.com/kelseyhightower/envconfig"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.22.0"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/ingka-group-digital/iam-proxy/internal/logger"
	"github.com/ingka-group-digital/iam-proxy/internal/version"
)

const (
	// ServiceName is the name of the service
	ServiceName = "iam-proxy"
)

// Config for iam-proxy service.
type Config struct {
	Host            string
	Port            int
	LogLevel        string
	IAM             IAM
	Environment     Environment
	HTTPTimeout     time.Duration
	ShutdownTimeout time.Duration
	Metric          Metric
	// Internal
	Logger *zap.SugaredLogger `ignored:"true"`
}

// Environment of kubernetes deployment.
type Environment string

// Valid environments
const (
	EnvDev   Environment = "development"
	EnvTest  Environment = "test"
	EnvStage Environment = "stage"
	EnvProd  Environment = "production"
)

// IAM defines the configuration for the iam auth2 functionalities
type IAM struct {
	Users  string
	Secret string
}

// Metric for OpenCensus trace and metric collection
type Metric struct {
	Enabled        bool
	OtelAgentAddr  string
	OpencensusAddr string
	// OpencensusReconnInterval represents the time interval, in seconds that
	// the OC Agent will attempt to reconnect.
	OpencensusReconnInterval int
	ReportingInterval        time.Duration
}

// New populates Config from environment variables and performs
// initializations in Config struct
func New() (*Config, error) {
	c := NewDefault()

	err := envconfig.Process("", &c)
	if err != nil {
		return nil, fmt.Errorf("parse application config %w", err)
	}

	// Initialize logger
	if err := c.initLogger(); err != nil {
		return nil, fmt.Errorf("initialize logger %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.HTTPTimeout)
	defer cancel()

	// Initialize tracer
	c.initTracer(ctx)

	return &c, nil
}

// NewDefault returns a Config with default values only
func NewDefault() Config {
	return Config{
		Port:            8080,
		LogLevel:        "info",
		HTTPTimeout:     5 * time.Second,
		ShutdownTimeout: 30 * time.Second,
	}
}

// newLogger initializes logger. In case of errors, it terminates the application
func (c *Config) initLogger() error {
	lc := zapdriver.NewProductionConfig()
	lc.InitialFields = map[string]interface{}{
		"version": version.Version,
	}

	switch c.LogLevel {
	case "debug":
		lc.Level.SetLevel(zap.DebugLevel)
	case "info":
		fallthrough
	default:
		lc.Level.SetLevel(zap.InfoLevel)
	}

	var err error
	logger.Logger, err = lc.Build()
	if err != nil {
		return err
	}

	c.Logger = logger.Logger.Sugar()

	return nil
}

// initTracer initializes OpenTelemetry tracer and OpenCensus metric collector
func (c *Config) initTracer(ctx context.Context) {
	if !c.Metric.Enabled {
		c.Logger.With("config", c.Metric).Info("Disabled opentelemetry tracer & opencensus metrics")
		return
	}
	c.Logger.With("config", c.Metric).Info("Enabled opentelemetry tracer & opencensus metrics")
	driver := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(c.Metric.OtelAgentAddr),
		otlptracegrpc.WithDialOption(grpc.WithBlock()), // useful for testing
	)

	exp, err := otlptrace.New(ctx, driver)

	if err != nil {
		c.Logger.Fatalw("Cannot connect to OpenTelemetry Exporter", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String(ServiceName),
		),
	)
	if err != nil {
		c.Logger.Fatalw("Cannot create otel resource", zap.Error(err))
	}

	bsp := sdktrace.NewBatchSpanProcessor(exp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// set global propagator to trace context (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetTracerProvider(tracerProvider)

	// TODO: Needs to be migrated to the latest OTEL version
	//cont := controller.basic.New(
	//	processor.basic.New(
	//		selector.simple.NewWithExactDistribution(),
	//		exp,
	//	),
	//	controller.basic.WithCollectPeriod(7*time.Second),
	//	controller.basic.WithExporter(exp),
	//)
	//global.SetMeterProvider(cont.MeterProvider())

	// Config the OpenCensus exporter.
	oc, err := ocagent.NewExporter(
		ocagent.WithInsecure(),
		ocagent.WithAddress(c.Metric.OpencensusAddr),
		ocagent.WithServiceName(ServiceName),
		ocagent.WithReconnectionPeriod(time.Second*time.Duration(c.Metric.OpencensusReconnInterval)),
	)
	if err != nil {
		c.Logger.Fatalw("Could not init OpenCensus exporter", zap.Error(err))
	}

	// This allows it to receive <stats> emitted by
	// OpenCensus-Go and upload them to the agent.
	view.RegisterExporter(oc)
	view.SetReportingPeriod(c.Metric.ReportingInterval)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
}
