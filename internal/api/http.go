// Copyright © 2024 Ingka Holding B.V. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

//go:generate gowrap gen -g -i Server -t ../../templates/opencensus_metrics.tpl -o http_server_with_metrics_gen.go

import (
	"context"
	"fmt"
	"net/http"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/ingka-group-digital/iam-proxy/client/health"
	"github.com/ingka-group-digital/iam-proxy/client/paths"
	"github.com/ingka-group-digital/iam-proxy/internal/config"
	"github.com/ingka-group-digital/iam-proxy/internal/logger"
	"github.com/ingka-group-digital/iam-proxy/internal/service"
)

// Server describes the methods needed to interact with this package interface.
type Server interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

// Config for HTTP Server
type Config struct {
	*config.Config
	Service    service.Servicer
	ListenAddr string
}

// Client for REST API
type Client struct {
	cfg    Config
	server http.Server
	tracer trace.Tracer
}

// New creates a new HTTP Server
func New(cfg Config) (*Client, error) {
	c := &Client{
		cfg:    cfg,
		tracer: otel.Tracer("api"),
	}

	router := c.setupRouter()

	if cfg.Metric.Enabled {
		if err := c.registerMetrics(); err != nil {
			return nil, fmt.Errorf("register metrics to opencensus: %w", err)
		}
	}

	var handler http.Handler = router

	c.server = http.Server{
		Addr:    cfg.ListenAddr,
		Handler: handler,
	}

	return c, nil
}

func (cl *Client) setupRouter() *gin.Engine {
	router := gin.New()

	switch cl.cfg.Environment {
	case config.EnvDev:
		gin.SetMode(gin.DebugMode)
	case config.EnvTest:
		gin.SetMode(gin.TestMode)
	case config.EnvStage:
		gin.SetMode(gin.TestMode)
	case config.EnvProd:
		gin.SetMode(gin.ReleaseMode)
	}

	l := cl.cfg.Logger.Desugar()

	// Group k8s endpoints to remove logs
	k8s := router.Group(paths.PathPrefix)
	{
		k8s.GET("/"+paths.HealthPath, cl.Health)
		k8s.GET("/"+paths.ReadyPath, cl.Ready)
		k8s.POST("/"+paths.OAuthToken, cl.Token)
		k8s.POST("/"+paths.ValidateToken, cl.Validate)
		k8s.POST("/"+paths.Identity, cl.Identity)
	}

	// Group /stocklevel-store/v1
	v1 := router.Group(paths.PathPrefix)
	if cl.cfg.Metric.Enabled {
		// Generate trace spans for HTTP requests
		v1.Use(otelgin.Middleware(config.ServiceName))
	}
	// Extract TraceID, create a logger instance with TraceID and add it in request context
	// Use logger.FromContext(c.Request.Context()) to get logger instance
	//v1.Use(correlation.AddTraceID)

	// Add a ginzap middleware, which:
	//   - Logs all requests, like a combined access and error log.
	//   - Logs to stdout.
	//   - RFC3339 with UTC time format.
	v1.Use(ginzap.Ginzap(l, time.RFC3339, true))

	// Logs all panic to error log
	//   - stack means whether output the stack info.
	v1.Use(ginzap.RecoveryWithZap(l, true))

	return router
}

func (cl *Client) registerMetrics() error {
	err := view.Register(
		ochttp.ServerRequestCountView,
		ochttp.ServerRequestBytesView,
		ochttp.ServerResponseBytesView,
		ochttp.ServerLatencyView,
		ochttp.ServerRequestCountByMethod,
		ochttp.ServerResponseCountByStatusCode,
		&view.View{
			Name:        "httpclient_latency_by_path",
			TagKeys:     []tag.Key{ochttp.KeyClientPath},
			Measure:     ochttp.ClientRoundtripLatency,
			Aggregation: ochttp.DefaultLatencyDistribution,
		},
	)

	return err
}

// ListenAndServe long-running process that listens and accepts incoming requests
func (cl *Client) ListenAndServe() error {
	return cl.server.ListenAndServe()
}

// Shutdown stops HTTP Server
func (cl *Client) Shutdown(ctx context.Context) error {
	return cl.server.Shutdown(ctx)
}

// Health swagger:route GET /health health
//
//	Responds to health checks.
//
//	Produces:
//	- application/json
//
//	Responses:
//	  200: body:health
//	  500:
//	  503:
func (cl *Client) Health(c *gin.Context) {
	log := logger.FromContext(c.Request.Context()).Sugar()
	h, err := cl.cfg.Service.Health(c.Request.Context())
	if err != nil {
		log.Errorw("Failed to check health", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	switch h.Status {
	case health.StatusAlive:
		c.JSON(http.StatusOK, h)
	case health.StatusDegraded:
		c.JSON(http.StatusInternalServerError, h)
	case health.StatusUnavailable:
		c.JSON(http.StatusServiceUnavailable, h)
	}
}

// Ready swagger:route GET /ready ready
//
//	Responds with 200 when service is ready to accept requests.
//
//	Responses:
//	  200:
//	  503:
func (cl *Client) Ready(c *gin.Context) {
	log := logger.FromContext(c.Request.Context()).Sugar()

	if err := cl.cfg.Service.Ready(c.Request.Context()); err != nil {
		log.Infow("Database is not ready", zap.Error(err))
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}

	c.Status(http.StatusOK)
}
