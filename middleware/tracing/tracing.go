package tracing

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Option is tracing option.
type Option func(*options)

type options struct {
	TracerProvider trace.TracerProvider
	Propagators    propagation.TextMapPropagator
}

// WithPropagators with tracer proagators.
func WithPropagators(propagators propagation.TextMapPropagator) Option {
	return func(opts *options) {
		opts.Propagators = propagators
	}
}

// WithTracerProvider with tracer privoder.
func WithTracerProvider(provider trace.TracerProvider) Option {
	return func(opts *options) {
		opts.TracerProvider = provider
	}
}

// Server returns a new server middleware for OpenTelemetry.
func Server(opts ...Option) middleware.Middleware {
	tracer := NewTracer(trace.SpanKindServer, opts...)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			var (
				component string
				operation string
				carrier   propagation.TextMapCarrier
			)
			tr, _ := transport.FromContext(ctx)
			info, _ := middleware.FromContext(ctx)
			operation = info.FullMethod
			component = string(tr.Kind)
			// TODO md carrier
			ctx, span := tracer.Start(ctx, component, operation, carrier)
			defer func() { tracer.End(ctx, span, err) }()
			return handler(ctx, req)
		}
	}
}

// Client returns a new client middleware for OpenTelemetry.
func Client(opts ...Option) middleware.Middleware {
	tracer := NewTracer(trace.SpanKindClient, opts...)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			var (
				component string
				operation string
				carrier   propagation.TextMapCarrier
			)
			tr, _ := transport.FromContext(ctx)
			info, _ := middleware.FromContext(ctx)
			component = string(tr.Kind)
			operation = info.FullMethod
			// TODO md carrier
			ctx, span := tracer.Start(ctx, component, operation, carrier)
			defer func() { tracer.End(ctx, span, err) }()
			return handler(ctx, req)
		}
	}
}
