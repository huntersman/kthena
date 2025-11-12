/*
Copyright The Volcano Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog/v2"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/volcano-sh/kthena/pkg/kthena-router/datastore"
	"github.com/volcano-sh/kthena/pkg/kthena-router/debug"
	"github.com/volcano-sh/kthena/pkg/kthena-router/router"
)

const (
	gracefulShutdownTimeout = 15 * time.Second
	routerConfigFile        = "/etc/config/routerConfiguration.yaml"
)

func NewRouter(store datastore.Store) *router.Router {
	return router.NewRouter(store, routerConfigFile)
}

// Starts router
func (s *Server) startRouter(ctx context.Context, router *router.Router, store datastore.Store) {
	gin.SetMode(gin.ReleaseMode)

	// Always start management endpoints on fixed port
	s.startManagementServer(ctx, router, store)

	// Get all Gateways from the store
	gatewayObjs := store.GetAllGateways()

	// If Gateways are configured, start multi-listener mode for /v1/*path
	if len(gatewayObjs) > 0 {
		// Convert to Gateway API types
		var gateways []*gatewayv1.Gateway
		for _, obj := range gatewayObjs {
			if gw, ok := obj.(*gatewayv1.Gateway); ok {
				gateways = append(gateways, gw)
			}
		}

		// Start multiple listeners based on Gateway configuration for /v1/*path only
		s.startMultiListeners(ctx, router, store, gateways)
	} else {
		klog.Info("No Gateways found, /v1/*path will be served on management port")
		// If no Gateways, /v1/*path is already handled by management server
	}
}

// startManagementServer starts the management HTTP server on fixed port
// This server handles healthz, readyz, metrics, debug endpoints, and /v1/*path (if no Gateways configured)
func (s *Server) startManagementServer(ctx context.Context, router *router.Router, store datastore.Store) {
	engine := gin.New()
	engine.Use(gin.LoggerWithWriter(gin.DefaultWriter, "/healthz", "/readyz", "/metrics"), gin.Recovery())

	// Management endpoints (no auth/access log middleware)
	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})

	engine.GET("/readyz", func(c *gin.Context) {
		if s.HasSynced() {
			c.JSON(http.StatusOK, gin.H{
				"message": "router is ready",
			})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"message": "router is not ready",
			})
		}
	})

	// Prometheus metrics endpoint
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Debug endpoints
	debugHandler := debug.NewDebugHandler(store)
	debugGroup := engine.Group("/debug/config_dump")
	{
		// List resources
		debugGroup.GET("/modelroutes", debugHandler.ListModelRoutes)
		debugGroup.GET("/modelservers", debugHandler.ListModelServers)
		debugGroup.GET("/pods", debugHandler.ListPods)

		// Get specific resources
		debugGroup.GET("/namespaces/:namespace/modelroutes/:name", debugHandler.GetModelRoute)
		debugGroup.GET("/namespaces/:namespace/modelservers/:name", debugHandler.GetModelServer)
		debugGroup.GET("/namespaces/:namespace/pods/:name", debugHandler.GetPod)
	}

	// Handle /v1/*path on management port only if no Gateways are configured
	// Add middleware for /v1/*path
	engine.Any("/v1/*path", AccessLogMiddleware(router), AuthMiddleware(router), router.HandlerFunc())

	server := &http.Server{
		Addr:    ":" + s.Port,
		Handler: engine.Handler(),
	}
	go func() {
		klog.Infof("Starting management server on port %s", s.Port)
		var err error
		if s.EnableTLS {
			if s.TLSCertFile == "" || s.TLSKeyFile == "" {
				klog.Fatalf("TLS enabled but cert or key file not specified")
			}
			err = server.ListenAndServeTLS(s.TLSCertFile, s.TLSKeyFile)
		} else {
			err = server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			klog.Fatalf("listen failed: %v", err)
		}
	}()

	go func() {
		<-ctx.Done()
		// graceful shutdown
		klog.Info("Shutting down management HTTP server ...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			klog.Errorf("Management server shutdown failed: %v", err)
		}
		klog.Info("Management HTTP server exited")
	}()
}

// startMultiListeners starts multiple HTTP servers based on Gateway listeners
// These servers only handle /v1/*path requests, management endpoints are on fixed port
func (s *Server) startMultiListeners(ctx context.Context, router *router.Router, store datastore.Store, gateways []*gatewayv1.Gateway) {
	var servers []*http.Server

	// Create a gin engine for each Gateway listener
	for _, gateway := range gateways {
		for _, listener := range gateway.Spec.Listeners {
			engine := gin.New()
			engine.Use(gin.Recovery())

			// Add middleware for /v1/*path only
			engine.Use(AccessLogMiddleware(router))
			engine.Use(AuthMiddleware(router))

			// Add hostname matching middleware if specified
			if listener.Hostname != nil && *listener.Hostname != "" {
				hostname := string(*listener.Hostname)
				engine.Use(func(c *gin.Context) {
					if c.Request.Host != hostname {
						c.AbortWithStatus(http.StatusNotFound)
						return
					}
					c.Next()
				})
			}

			// Only handle /v1/*path on Gateway listeners
			engine.Any("/v1/*path", router.HandlerFunc())

			// Return 404 for all other paths (management endpoints are on fixed port)
			engine.NoRoute(func(c *gin.Context) {
				c.JSON(http.StatusNotFound, gin.H{
					"message": "Not found. Management endpoints are available on the management port.",
				})
			})

			port := int32(listener.Port)
			server := &http.Server{
				Addr:    ":" + strconv.Itoa(int(port)),
				Handler: engine.Handler(),
			}

			servers = append(servers, server)

			listenerName := string(listener.Name)
			protocol := string(listener.Protocol)
			go func(name string, p int32, proto string, srv *http.Server) {
				klog.Infof("Starting Gateway listener %s on port %d with protocol %s for /v1/*path", name, p, proto)
				var err error
				if proto == string(gatewayv1.HTTPSProtocolType) {
					// For HTTPS, we need to get TLS cert/key from the listener's TLS config
					// For simplicity, we'll use the server's TLS config if available
					if s.EnableTLS && s.TLSCertFile != "" && s.TLSKeyFile != "" {
						err = srv.ListenAndServeTLS(s.TLSCertFile, s.TLSKeyFile)
					} else {
						klog.Errorf("HTTPS listener %s requires TLS configuration", name)
						return
					}
				} else {
					err = srv.ListenAndServe()
				}
				if err != nil && err != http.ErrServerClosed {
					klog.Fatalf("listen failed for listener %s: %v", name, err)
				}
			}(listenerName, port, protocol, server)
		}
	}

	go func() {
		<-ctx.Done()
		// graceful shutdown
		klog.Info("Shutting down Gateway listener servers ...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
		defer cancel()
		for _, server := range servers {
			if err := server.Shutdown(shutdownCtx); err != nil {
				klog.Errorf("Gateway listener server shutdown failed: %v", err)
			}
		}
		klog.Info("All Gateway listener servers exited")
	}()
}

func AccessLogMiddleware(gwRouter *router.Router) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Access log for "/v1/" only
		if !strings.HasPrefix(c.Request.URL.Path, "/v1/") {
			c.Next()
			return
		}

		// Calling Middleware
		gwRouter.AccessLog()(c)
	}
}

func AuthMiddleware(gwRouter *router.Router) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Auth for "/v1/" only
		if !strings.HasPrefix(c.Request.URL.Path, "/v1/") {
			c.Next()
			return
		}

		// Calling Middleware
		gwRouter.Auth()(c)
		if c.IsAborted() {
			return
		}

		c.Next()
	}
}
