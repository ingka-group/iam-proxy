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

package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/ingka-group-digital/iam-proxy/internal/api"
	"github.com/ingka-group-digital/iam-proxy/internal/config"
	"github.com/ingka-group-digital/iam-proxy/internal/errors"
	"github.com/ingka-group-digital/iam-proxy/internal/service"
)

// Run starts the application
func Run() error {
	c, err := config.New()
	if err != nil {
		return fmt.Errorf("failed to initialise application config: %w", err)
	}

	c.Logger.Debug("Creating iam service")
	svc, err := service.New(service.Config{
		Config: c,
	})
	if err != nil {
		c.Logger.Errorw("Failed to create iam-proxy service", zap.Error(err))
		return err
	}

	c.Logger.Debug("Creating HTTP Server")
	srv, err := api.New(api.Config{
		ListenAddr: fmt.Sprintf("%s:%d", c.Host, c.Port),
		Config:     c,
		Service:    service.NewWithTracing(service.NewWithMetrics(svc, "app"), "app"),
	})
	if err != nil {
		c.Logger.Errorw("Failed to create HTTP Server", zap.Error(err))
		return err
	}
	srvWrapped := api.NewWithMetrics(srv, "api")

	stop := make(chan os.Signal, 1)

	// interrupt signal sent from terminal
	signal.Notify(stop, os.Interrupt)
	// sigterm signal sent from kubernetes
	signal.Notify(stop, syscall.SIGTERM)

	go func() {
		c.Logger.Infow("HTTP Server listening",
			"host", c.Host,
			"port", c.Port)
		if err := srvWrapped.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				c.Logger.Errorw("HTTP Server stopped unexpectedly", zap.Error(err))
				stop <- os.Interrupt
			}
		}
	}()

	<-stop

	c.Logger.Infow("Shutting down",
		"service", config.ServiceName,
		"timeout", c.ShutdownTimeout,
	)

	ctx, cancel := context.WithTimeout(context.Background(), c.ShutdownTimeout)
	defer cancel()

	cerr := make(chan error, 1)

	go func() {
		c.Logger.Info("Closing HTTP server")

		var errs errors.Errors

		if err = srvWrapped.Shutdown(ctx); err != nil {
			c.Logger.Errorw("Failed to stop HTTP server", zap.Error(err))
			errs = append(errs, err)
		}

		cerr <- errs.Join()
	}()

	select {
	case <-ctx.Done():
		c.Logger.Errorw("Service is not shut down within shutdown timeout", zap.Error(ctx.Err()))
		return ctx.Err()
	case err := <-cerr:
		if err != nil {
			c.Logger.Errorw("Could not stop service gracefully", zap.Error(err))
			return err
		}
		c.Logger.Infof("%s is shut down", config.ServiceName)
		return nil
	}
}
