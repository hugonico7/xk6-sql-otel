package sql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func init() {
	sql.Register(oracleDriverName, stubOracleDriver{})
}

type stubOracleDriver struct{}

func (stubOracleDriver) Open(string) (driver.Conn, error) {
	return stubOracleConn{}, nil
}

type stubOracleConn struct{}

func (stubOracleConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}

func (stubOracleConn) Close() error {
	return nil
}

func (stubOracleConn) Begin() (driver.Tx, error) {
	return nil, errors.New("not implemented")
}

func TestOpenDriverName(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://127.0.0.1:4318")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")

	resetOracleOTelDriver()

	wrappedName, err := openDriverName(oracleDriverName)
	require.NoError(t, err)
	require.NotEmpty(t, wrappedName)
	require.NotEqual(t, oracleDriverName, wrappedName)

	again, err := openDriverName(oracleDriverName)
	require.NoError(t, err)
	require.Equal(t, wrappedName, again)

	passthrough, err := openDriverName("ramsql")
	require.NoError(t, err)
	require.Equal(t, "ramsql", passthrough)
}

func TestOpenDriverNameWithoutExporter(t *testing.T) {
	resetOracleOTelDriver()

	wrappedName, err := openDriverName(oracleDriverName)
	require.NoError(t, err)
	require.Equal(t, oracleDriverName, wrappedName)
}

func TestFlushOracleTracerProviderWithoutProvider(t *testing.T) {
	resetOracleOTelDriver()

	require.NoError(t, flushOracleTracerProvider(context.Background()))
}

func resetOracleOTelDriver() {
	oracleOTelDriver.mu.Lock()
	defer oracleOTelDriver.mu.Unlock()

	oracleOTelDriver.name = ""
	oracleOTelDriver.provider = nil
	oracleOTelDriver.configured = false
}
