package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	shutdown := initTracer()
	defer shutdown()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	r.With(otelMiddleware).Get("/hello", helloHandler)

	log.Println("Starting chi server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	span.AddEvent("processing /hello")

	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprintln(w, "Hello from OTLP-traced chi server!")
	if err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func otelMiddleware(next http.Handler) http.Handler {
	return otelhttp.NewHandler(next, "http-request")
}

func initTracer() func() {
	ctx := context.Background()

	// This env var may contain http://, so we sanitize it
	raw := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if raw == "" {
		raw = "http://jaeger-collector.observability:4318"
	}

	// Strip "http://" if present, as WithEndpoint expects host:port only
	endpoint := raw
	if len(raw) > 7 && raw[:7] == "http://" {
		endpoint = raw[7:]
	}

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint), // <-- host:port only
		otlptracehttp.WithInsecure(),         // <-- because we're using HTTP
		otlptracehttp.WithTimeout(5*time.Second),
	)
	if err != nil {
		log.Fatalf("failed to create OTLP HTTP exporter: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("chi-apiserver"),
		)),
	)

	otel.SetTracerProvider(tp)

	return func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatalf("error shutting down tracer provider: %v", err)
		}
	}
}
