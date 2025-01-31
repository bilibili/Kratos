package consul

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/consul/api"

	"github.com/go-kratos/kratos/v2/registry"
)

var (
	_ registry.Registrar = (*Registry)(nil)
	_ registry.Discovery = (*Registry)(nil)
)

// Option is consul registry option.
type Option func(*Registry)

// WithHealthCheck with registry health check option.
func WithHealthCheck(enable bool) Option {
	return func(o *Registry) {
		o.enableHealthCheck = enable
	}
}

// WithTimeout with get services timeout option.
func WithTimeout(timeout time.Duration) Option {
	return func(o *Registry) {
		o.timeout = timeout
	}
}

// WithDatacenter with registry datacenter option
func WithDatacenter(dc Datacenter) Option {
	return func(o *Registry) {
		o.cli.dc = dc
	}
}

// WithHeartbeat enable or disable heartbeat
func WithHeartbeat(enable bool) Option {
	return func(o *Registry) {
		if o.cli != nil {
			o.cli.heartbeat = enable
		}
	}
}

// WithServiceResolver with endpoint function option.
func WithServiceResolver(fn ServiceResolver) Option {
	return func(o *Registry) {
		if o.cli != nil {
			o.cli.resolver = fn
		}
	}
}

// WithHealthCheckInterval with healthcheck interval in seconds.
func WithHealthCheckInterval(interval int) Option {
	return func(o *Registry) {
		if o.cli != nil {
			o.cli.healthcheckInterval = interval
		}
	}
}

// WithDeregisterCriticalServiceAfter with deregister-critical-service-after in seconds.
func WithDeregisterCriticalServiceAfter(interval int) Option {
	return func(o *Registry) {
		if o.cli != nil {
			o.cli.deregisterCriticalServiceAfter = interval
		}
	}
}

// WithServiceCheck with service checks
func WithServiceCheck(checks ...*api.AgentServiceCheck) Option {
	return func(o *Registry) {
		if o.cli != nil {
			o.cli.serviceChecks = checks
		}
	}
}

// Config is consul registry config
type Config struct {
	*api.Config
}

// Registry is consul registry
type Registry struct {
	cli               *Client
	enableHealthCheck bool
	registry          map[string]*serviceSet
	lock              sync.RWMutex
	timeout           time.Duration
}

// New creates consul registry
func New(apiClient *api.Client, opts ...Option) *Registry {
	r := &Registry{
		registry:          make(map[string]*serviceSet),
		enableHealthCheck: true,
		timeout:           10 * time.Second,
		cli: &Client{
			dc:                             SingleDatacenter,
			cli:                            apiClient,
			resolver:                       defaultResolver,
			healthcheckInterval:            10,
			heartbeat:                      true,
			deregisterCriticalServiceAfter: 600,
			cancelers:                      make(map[string]*canceler),
		},
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Register register service
func (r *Registry) Register(ctx context.Context, svc *registry.ServiceInstance) error {
	return r.cli.Register(ctx, svc, r.enableHealthCheck)
}

// Deregister deregister service
func (r *Registry) Deregister(ctx context.Context, svc *registry.ServiceInstance) error {
	return r.cli.Deregister(ctx, svc.ID)
}

// GetService return service by name
func (r *Registry) GetService(ctx context.Context, name string) ([]*registry.ServiceInstance, error) {
	r.lock.RLock()
	set := r.registry[name]
	r.lock.RUnlock()

	getRemote := func() []*registry.ServiceInstance {
		services, _, err := r.cli.Service(ctx, name, 0, true)
		if err == nil && len(services) > 0 {
			return services
		}
		return nil
	}

	if set == nil {
		if s := getRemote(); len(s) > 0 {
			return s, nil
		}
		return nil, fmt.Errorf("service %s not resolved in registry", name)
	}
	ss, _ := set.services.Load().([]*registry.ServiceInstance)
	if ss == nil {
		if s := getRemote(); len(s) > 0 {
			return s, nil
		}
		return nil, fmt.Errorf("service %s not found in registry", name)
	}
	return ss, nil
}

// ListServices return service list.
func (r *Registry) ListServices() (allServices map[string][]*registry.ServiceInstance, err error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	allServices = make(map[string][]*registry.ServiceInstance)
	for name, set := range r.registry {
		var services []*registry.ServiceInstance
		ss, _ := set.services.Load().([]*registry.ServiceInstance)
		if ss == nil {
			continue
		}
		services = append(services, ss...)
		allServices[name] = services
	}
	return
}

// Watch resolve service by name
func (r *Registry) Watch(ctx context.Context, name string) (registry.Watcher, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.lock.Lock()
	set, ok := r.registry[name]
	if !ok {
		cancelCtx, cancel := context.WithCancel(context.Background())
		set = &serviceSet{
			registry:    r,
			watcher:     make(map[*watcher]struct{}),
			services:    &atomic.Value{},
			serviceName: name,
			ctx:         cancelCtx,
			cancel:      cancel,
		}
		r.registry[name] = set
	}
	set.ref.Add(1)
	r.lock.Unlock()

	// init watcher
	w := &watcher{
		event: make(chan struct{}, 1),
	}
	w.ctx, w.cancel = context.WithCancel(ctx)
	w.set = set
	set.lock.Lock()
	set.watcher[w] = struct{}{}
	set.lock.Unlock()

	ss, _ := set.services.Load().([]*registry.ServiceInstance)
	if len(ss) > 0 {
		// If the service has a value, it needs to be pushed to the watcher,
		// otherwise the initial data may be blocked forever during the watch.
		select {
		case w.event <- struct{}{}:
		default:
		}
	}

	if !ok {
		if err := r.resolve(ctx, set); err != nil {
			return nil, err
		}
	}
	return w, nil
}

func (r *Registry) resolve(ctx context.Context, ss *serviceSet) error {
	listServices := r.cli.Service
	if r.timeout > 0 {
		listServices = func(ctx context.Context, service string, index uint64, passingOnly bool) ([]*registry.ServiceInstance, uint64, error) {
			timeoutCtx, cancel := context.WithTimeout(ctx, r.timeout)
			defer cancel()

			return r.cli.Service(timeoutCtx, service, index, passingOnly)
		}
	}

	services, idx, err := listServices(ctx, ss.serviceName, 0, true)
	if err != nil {
		return err
	}
	if len(services) > 0 {
		ss.broadcast(services)
	}

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				tmpService, tmpIdx, err := listServices(ss.ctx, ss.serviceName, idx, true)
				if err != nil {
					if err := sleepCtx(ss.ctx, time.Second); err != nil {
						return
					}
					continue
				}
				if tmpIdx != idx {
					services = tmpService
					ss.broadcast(services)
				}
				idx = tmpIdx
			case <-ss.ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (r *Registry) tryDelete(ss *serviceSet) bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	if ss.ref.Add(-1) != 0 {
		return false
	}
	ss.cancel()
	delete(r.registry, ss.serviceName)
	return true
}
