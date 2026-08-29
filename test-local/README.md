# Local Oracle + Jaeger test

## Start dependencies

```bash
docker compose -f ./test-local/docker-compose.yml up -d
```

Jaeger UI: `http://localhost:16686`

Oracle connection string used by the test script:

```text
oracle://app:app@localhost:1521/FREEPDB1
```

## Build local k6

Install `xk6` if needed:

```bash
go install go.k6.io/xk6/cmd/xk6@latest
```

Build a local binary into `test-local/k6` from the repo root:

```bash
"$HOME/go/bin/xk6" build \
  --output "./test-local/k6" \
  --with github.com/grafana/xk6-sql=. \
  --with github.com/denyshuzovskyi/xk6-sql-driver-oracle@latest
```

## Run the test

This fork exports SQL traces directly through OpenTelemetry HTTP. For local validation against Jaeger:

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 \
OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf \
OTEL_SERVICE_NAME=xk6-sql-local \
./test-local/k6 run ./test-local/script.js
```

For Dynatrace, use the same command shape with the Dynatrace base OTLP endpoint and headers:

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=https://<environment>.live.dynatrace.com/api/v2/otlp \
OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf \
OTEL_EXPORTER_OTLP_HEADERS='Authorization=Api-Token <TOKEN>' \
OTEL_SERVICE_NAME=xk6-sql \
./test-local/k6 run ./test-local/script.js
```

## What to verify

- The script finishes successfully.
- Jaeger shows a service for the run.
- SQL spans are present.
- `db.system.name` is `oracle.db`.
