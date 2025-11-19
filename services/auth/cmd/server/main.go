package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/ikampus/auth-service/internal/config"
	"github.com/ikampus/auth-service/internal/handler"
	"github.com/ikampus/auth-service/internal/service"
	pb "github.com/ikampus/protos/auth"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("Starting AuthService...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	logger.WithFields(logrus.Fields{
		"environment": cfg.Server.Environment,
		"grpc_port":   cfg.Server.GRPCPort,
		"http_port":   cfg.Server.Port,
	}).Info("Configuration loaded")

	// Connect to database
	db, err := connectDatabase(cfg.Database, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	logger.Info("Database connection established")

	// Create service
	authService, err := service.NewAuthService(db, cfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create auth service")
	}

	// Create gRPC handler
	authHandler := handler.NewAuthHandler(authService, logger)

	// Create gRPC server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor(logger)),
	)

	// Register services
	pb.RegisterAuthServiceServer(grpcServer, authHandler)

	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("ikampus.auth.AuthService", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection service (for grpcurl and development)
	reflection.Register(grpcServer)

	// Start gRPC server
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
	if err != nil {
		logger.WithError(err).Fatal("Failed to create gRPC listener")
	}

	// Start HTTP health check server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: createHTTPHandler(logger, db),
	}

	// Start servers in goroutines
	go func() {
		logger.WithField("port", cfg.Server.GRPCPort).Info("Starting gRPC server")
		if err := grpcServer.Serve(grpcListener); err != nil {
			logger.WithError(err).Fatal("Failed to serve gRPC")
		}
	}()

	go func() {
		logger.WithField("port", cfg.Server.Port).Info("Starting HTTP health check server")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Error("HTTP server error")
		}
	}()

	logger.Info("AuthService is running")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down servers...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Stop accepting new requests
	healthServer.SetServingStatus("ikampus.auth.AuthService", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	// Shutdown HTTP server
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Error("HTTP server shutdown error")
	}

	// Gracefully stop gRPC server
	grpcServer.GracefulStop()

	logger.Info("Servers stopped gracefully")
}

// connectDatabase establishes database connection with retry logic
func connectDatabase(cfg config.DatabaseConfig, logger *logrus.Logger) (*sql.DB, error) {
	dsn := cfg.DatabaseDSN()

	var db *sql.DB
	var err error

	// Retry connection up to 5 times
	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			logger.WithError(err).Warnf("Failed to open database connection (attempt %d/5)", i+1)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		// Ping to verify connection
		if err = db.Ping(); err != nil {
			logger.WithError(err).Warnf("Failed to ping database (attempt %d/5)", i+1)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		// Connection successful
		break
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after 5 attempts: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return db, nil
}

// createHTTPHandler creates HTTP handler for health checks and metrics
func createHTTPHandler(logger *logrus.Logger, db *sql.DB) http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Check database connection
		if err := db.Ping(); err != nil {
			logger.WithError(err).Error("Database health check failed")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unhealthy","database":"down"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","database":"up"}`))
	})

	// Readiness check endpoint
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		// Check database connection
		if err := db.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"ready":false}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ready":true}`))
	})

	// Liveness check endpoint
	mux.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"alive":true}`))
	})

	// Metrics endpoint (placeholder for Prometheus metrics)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Add Prometheus metrics
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Metrics coming soon\n"))
	})

	return mux
}

// loggingInterceptor logs all gRPC requests
func loggingInterceptor(logger *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Call the handler
		resp, err := handler(ctx, req)

		// Log the request
		duration := time.Since(start)
		fields := logrus.Fields{
			"method":   info.FullMethod,
			"duration": duration.String(),
		}

		if err != nil {
			fields["error"] = err.Error()
			logger.WithFields(fields).Error("gRPC request failed")
		} else {
			logger.WithFields(fields).Info("gRPC request completed")
		}

		return resp, err
	}
}
