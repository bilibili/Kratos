module github.com/go-kratos/kratos/middleware/tracing/v2

go 1.15

require (
	github.com/go-kratos/kratos/v2 v2.0.5
	go.opentelemetry.io/otel v1.0.0-RC1
	go.opentelemetry.io/otel/sdk v1.0.0-RC1
	go.opentelemetry.io/otel/trace v1.0.0-RC1
)

replace github.com/go-kratos/kratos/v2 => ../../../kratos
