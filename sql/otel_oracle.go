package sql

import (
	"context"
	"os"
	"sync"

	"github.com/XSAM/otelsql"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const oracleDriverName = "oracle"
const defaultServiceName = "k6-sql"

// oracleOTelDriver keeps the Oracle-specific otelsql wrapper and exporter state.
// The exporter is optional and is only initialized when standard OTEL env vars are present.
var oracleOTelDriver struct { //nolint:gochecknoglobals
	mu         sync.Mutex
	name       string
	provider   *sdktrace.TracerProvider
	configured bool
}

func openDriverName(driverName string) (string, error) {
	if driverName != oracleDriverName {
		return driverName, nil
	}

	oracleOTelDriver.mu.Lock()
	defer oracleOTelDriver.mu.Unlock()

	provider, err := oracleTracerProvider()
	if err != nil {
		return "", err
	}
	if provider == nil {
		return driverName, nil
	}

	if oracleOTelDriver.name != "" {
		return oracleOTelDriver.name, nil
	}

	wrappedName, err := otelsql.Register(
		driverName,
		otelsql.WithTracerProvider(provider),
		otelsql.WithAttributes(attribute.String("db.system.name", "oracle.db")),
		otelsql.WithSpanOptions(otelsql.SpanOptions{DisableQuery: true}),
	)
	if err != nil {
		return "", err
	}

	oracleOTelDriver.name = wrappedName

	return oracleOTelDriver.name, nil
}

// oracleTracerProvider returns a process-wide provider configured through OTEL_* variables.
// If no OTLP endpoint is configured, Oracle falls back to the unwrapped driver.
func oracleTracerProvider() (*sdktrace.TracerProvider, error) {
	if oracleOTelDriver.configured {
		return oracleOTelDriver.provider, nil
	}

	oracleOTelDriver.configured = true

	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" && os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") == "" {
		return nil, nil
	}

	exporter, err := otlptracehttp.New(context.Background())
	if err != nil {
		return nil, err
	}

	oracleOTelDriver.provider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewSchemaless(attribute.String("service.name", serviceName()))),
	)

	return oracleOTelDriver.provider, nil
}

func flushOracleTracerProvider(ctx context.Context) error {
	oracleOTelDriver.mu.Lock()
	provider := oracleOTelDriver.provider
	oracleOTelDriver.mu.Unlock()

	if provider == nil {
		return nil
	}

	return provider.ForceFlush(ctx)
}

func serviceName() string {
	if value := os.Getenv("OTEL_SERVICE_NAME"); value != "" {
		return value
	}

	return defaultServiceName
}
