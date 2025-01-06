package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpcinsecure "google.golang.org/grpc/credentials/insecure"
	grpcmd "google.golang.org/grpc/metadata"

	"github.com/go-kratos/kratos/v2/internal/matcher"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/selector/wrr"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc/resolver/discovery"

	// init resolver
	_ "github.com/go-kratos/kratos/v2/transport/grpc/resolver/direct"
)

func init() {
	if selector.GlobalSelector() == nil {
		selector.SetGlobalSelector(wrr.NewBuilder())
	}
}

// ClientOption is gRPC client option.
type ClientOption func(o *clientOptions)

// WithEndpoint with client endpoint.
func WithEndpoint(endpoint string) ClientOption {
	return func(o *clientOptions) {
		o.endpoint = endpoint
	}
}

// WithSubset with client discovery subset size.
// zero value means subset filter disabled
func WithSubset(size int) ClientOption {
	return func(o *clientOptions) {
		o.subsetSize = size
	}
}

// WithTimeout with client timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.timeout = timeout
	}
}

// WithMiddleware with client middleware.
func WithMiddleware(m ...middleware.Middleware) ClientOption {
	return func(o *clientOptions) {
		o.middleware = m
	}
}

// WithStreamMiddleware with client stream middleware.
func WithStreamMiddleware(m ...middleware.Middleware) ClientOption {
	return func(s *clientOptions) {
		s.streamMiddleware = m
	}
}

// WithDiscovery with client discovery.
func WithDiscovery(d registry.Discovery) ClientOption {
	return func(o *clientOptions) {
		o.discovery = d
	}
}

// WithTLSConfig with TLS config.
func WithTLSConfig(c *tls.Config) ClientOption {
	return func(o *clientOptions) {
		o.tlsConf = c
	}
}

// WithUnaryInterceptor returns a DialOption that specifies the interceptor for unary RPCs.
func WithUnaryInterceptor(in ...grpc.UnaryClientInterceptor) ClientOption {
	return func(o *clientOptions) {
		o.ints = in
	}
}

// WithStreamInterceptor returns a DialOption that specifies the interceptor for streaming RPCs.
func WithStreamInterceptor(in ...grpc.StreamClientInterceptor) ClientOption {
	return func(o *clientOptions) {
		o.streamInts = in
	}
}

// WithOptions with gRPC options.
func WithOptions(opts ...grpc.DialOption) ClientOption {
	return func(o *clientOptions) {
		o.grpcOpts = opts
	}
}

// WithNodeFilter with select filters
func WithNodeFilter(filters ...selector.NodeFilter) ClientOption {
	return func(o *clientOptions) {
		o.filters = filters
	}
}

// WithHealthCheck with health check
func WithHealthCheck(healthCheck bool) ClientOption {
	return func(o *clientOptions) {
		if !healthCheck {
			o.healthCheckConfig = ""
		}
	}
}

// WithLogger with logger
// Deprecated: use global logger instead.
func WithLogger(log.Logger) ClientOption {
	return func(*clientOptions) {}
}

func WithPrintDiscoveryDebugLog(p bool) ClientOption {
	return func(o *clientOptions) {
		o.printDiscoveryDebugLog = p
	}
}

// clientOptions is gRPC Client
type clientOptions struct {
	endpoint               string
	subsetSize             int
	tlsConf                *tls.Config
	timeout                time.Duration
	discovery              registry.Discovery
	middleware             []middleware.Middleware
	streamMiddleware       []middleware.Middleware
	ints                   []grpc.UnaryClientInterceptor
	streamInts             []grpc.StreamClientInterceptor
	grpcOpts               []grpc.DialOption
	balancerName           string
	filters                []selector.NodeFilter
	healthCheckConfig      string
	printDiscoveryDebugLog bool
}

// Dial returns a GRPC connection.
func Dial(ctx context.Context, opts ...ClientOption) (*grpc.ClientConn, error) {
	return dial(ctx, false, opts...)
}

// DialInsecure returns an insecure GRPC connection.
func DialInsecure(ctx context.Context, opts ...ClientOption) (*grpc.ClientConn, error) {
	return dial(ctx, true, opts...)
}

func dial(ctx context.Context, insecure bool, opts ...ClientOption) (*grpc.ClientConn, error) {
	options := clientOptions{
		timeout:                2000 * time.Millisecond,
		balancerName:           balancerName,
		subsetSize:             25,
		printDiscoveryDebugLog: true,
		healthCheckConfig:      `,"healthCheckConfig":{"serviceName":""}`,
	}
	for _, o := range opts {
		o(&options)
	}
	ints := []grpc.UnaryClientInterceptor{
		unaryClientInterceptor(options.middleware, options.timeout, options.filters),
	}
	sints := []grpc.StreamClientInterceptor{
		streamClientInterceptor(options.streamMiddleware, options.filters),
	}

	if len(options.ints) > 0 {
		ints = append(ints, options.ints...)
	}
	if len(options.streamInts) > 0 {
		sints = append(sints, options.streamInts...)
	}
	grpcOpts := []grpc.DialOption{
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingConfig": [{"%s":{}}]%s}`,
			options.balancerName, options.healthCheckConfig)),
		grpc.WithChainUnaryInterceptor(ints...),
		grpc.WithChainStreamInterceptor(sints...),
	}

	if options.discovery != nil {
		grpcOpts = append(grpcOpts,
			grpc.WithResolvers(
				discovery.NewBuilder(
					options.discovery,
					discovery.WithInsecure(insecure),
					discovery.WithTimeout(options.timeout),
					discovery.WithSubset(options.subsetSize),
					discovery.PrintDebugLog(options.printDiscoveryDebugLog),
				)))
	}
	if insecure {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(grpcinsecure.NewCredentials()))
	}
	if options.tlsConf != nil {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(credentials.NewTLS(options.tlsConf)))
	}
	if len(options.grpcOpts) > 0 {
		grpcOpts = append(grpcOpts, options.grpcOpts...)
	}
	return grpc.DialContext(ctx, options.endpoint, grpcOpts...)
}

func unaryClientInterceptor(ms []middleware.Middleware, timeout time.Duration, filters []selector.NodeFilter) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = transport.NewClientContext(ctx, &Transport{
			endpoint:    cc.Target(),
			operation:   method,
			reqHeader:   headerCarrier{},
			nodeFilters: filters,
		})
		if timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
		h := func(ctx context.Context, req interface{}) (interface{}, error) {
			if tr, ok := transport.FromClientContext(ctx); ok {
				header := tr.RequestHeader()
				keys := header.Keys()
				keyvals := make([]string, 0, len(keys))
				for _, k := range keys {
					keyvals = append(keyvals, k, header.Get(k))
				}
				ctx = grpcmd.AppendToOutgoingContext(ctx, keyvals...)
			}
			return reply, invoker(ctx, method, req, reply, cc, opts...)
		}
		if len(ms) > 0 {
			h = middleware.Chain(ms...)(h)
		}
		var p selector.Peer
		ctx = selector.NewPeerContext(ctx, &p)
		_, err := h(ctx, req)
		return err
	}
}

// wrappedClientStream wraps the grpc.ClientStream and applies middleware
type wrappedClientStream struct {
	grpc.ClientStream
	ctx        context.Context
	middleware matcher.Matcher
}

func (w *wrappedClientStream) Context() context.Context {
	return w.ctx
}

func (w *wrappedClientStream) SendMsg(m interface{}) error {
	h := func(_ context.Context, req interface{}) (interface{}, error) {
		return req, w.ClientStream.SendMsg(m)
	}

	info, ok := transport.FromClientContext(w.ctx)
	if !ok {
		return fmt.Errorf("transport value stored in ctx returns: %v", ok)
	}

	if next := w.middleware.Match(info.Operation()); len(next) > 0 {
		h = middleware.Chain(next...)(h)
	}

	_, err := h(w.ctx, m)
	return err
}

func (w *wrappedClientStream) RecvMsg(m interface{}) error {
	h := func(_ context.Context, req interface{}) (interface{}, error) {
		return req, w.ClientStream.RecvMsg(m)
	}

	info, ok := transport.FromClientContext(w.ctx)
	if !ok {
		return fmt.Errorf("transport value stored in ctx returns: %v", ok)
	}

	if next := w.middleware.Match(info.Operation()); len(next) > 0 {
		h = middleware.Chain(next...)(h)
	}

	_, err := h(w.ctx, m)
	return err
}

func streamClientInterceptor(ms []middleware.Middleware, filters []selector.NodeFilter) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) { // nolint
		ctx = transport.NewClientContext(ctx, &Transport{
			endpoint:    cc.Target(),
			operation:   method,
			reqHeader:   headerCarrier{},
			nodeFilters: filters,
		})
		var p selector.Peer
		ctx = selector.NewPeerContext(ctx, &p)

		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			return nil, err
		}

		h := func(_ context.Context, _ interface{}) (interface{}, error) {
			return streamer, nil
		}

		m := matcher.New()
		if len(ms) > 0 {
			m.Use(ms...)
			middleware.Chain(ms...)(h)
		}

		wrappedStream := &wrappedClientStream{
			ClientStream: clientStream,
			ctx:          ctx,
			middleware:   m,
		}

		return wrappedStream, nil
	}
}
