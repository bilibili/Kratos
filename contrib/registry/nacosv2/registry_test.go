package nacos

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"

	"github.com/go-kratos/kratos/v2/registry"
)

func TestRegistry_Register(t *testing.T) {
	sc := []constant.ServerConfig{
		*constant.NewServerConfig("127.0.0.1", 8848),
	}

	cc := constant.ClientConfig{
		NamespaceId:         "public", // namespace id
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "debug",
	}

	// a more graceful way to create naming client
	client, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	r := New(client)

	testServer := &registry.ServiceInstance{
		ID:        "1",
		Name:      "test1",
		Version:   "v1.0.0",
		Endpoints: []string{"http://127.0.0.1:8080?isSecure=false"},
	}
	testServerWithMetadata := &registry.ServiceInstance{
		ID:        "1",
		Name:      "test1",
		Version:   "v1.0.0",
		Endpoints: []string{"http://127.0.0.1:8080?isSecure=false"},
		Metadata:  map[string]string{"idc": "shanghai-xs"},
	}
	type fields struct {
		registry *Registry
	}
	type args struct {
		ctx     context.Context
		service *registry.ServiceInstance
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		deferFunc func(t *testing.T)
	}{
		{
			name: "normal",
			fields: fields{
				registry: New(client),
			},
			args: args{
				ctx:     context.Background(),
				service: testServer,
			},
			wantErr: false,
			deferFunc: func(t *testing.T) {
				err = r.Deregister(context.Background(), testServer)
				if err != nil {
					t.Error(err)
				}
			},
		},
		{
			name: "withMetadata",
			fields: fields{
				registry: New(client),
			},
			args: args{
				ctx:     context.Background(),
				service: testServerWithMetadata,
			},
			wantErr: false,
			deferFunc: func(t *testing.T) {
				err = r.Deregister(context.Background(), testServerWithMetadata)
				if err != nil {
					t.Error(err)
				}
			},
		},
		{
			name: "error",
			fields: fields{
				registry: New(client),
			},
			args: args{
				ctx: context.Background(),
				service: &registry.ServiceInstance{
					ID:        "1",
					Name:      "",
					Version:   "v1.0.0",
					Endpoints: []string{"http://127.0.0.1:8080?isSecure=false"},
				},
			},
			wantErr: true,
		},
		{
			name: "urlError",
			fields: fields{
				registry: New(client),
			},
			args: args{
				ctx: context.Background(),
				service: &registry.ServiceInstance{
					ID:        "1",
					Name:      "test",
					Version:   "v1.0.0",
					Endpoints: []string{"127.0.0.1:8080"},
				},
			},
			wantErr: true,
		},
		{
			name: "portError",
			fields: fields{
				registry: New(client),
			},
			args: args{
				ctx: context.Background(),
				service: &registry.ServiceInstance{
					ID:        "1",
					Name:      "test",
					Version:   "v1.0.0",
					Endpoints: []string{"http://127.0.0.1888"},
				},
			},
			wantErr: true,
		},
		{
			name: "withCluster",
			fields: fields{
				registry: New(client, WithCluster("test")),
			},
			args: args{
				ctx:     context.Background(),
				service: testServer,
			},
			wantErr: false,
		},
		{
			name: "withGroup",
			fields: fields{
				registry: New(client, WithGroup("TEST_GROUP")),
			},
			args: args{
				ctx:     context.Background(),
				service: testServer,
			},
			wantErr: false,
		},
		{
			name: "withWeight",
			fields: fields{
				registry: New(client, WithWeight(200)),
			},
			args: args{
				ctx:     context.Background(),
				service: testServer,
			},
			wantErr: false,
		},
		{
			name: "withPrefix",
			fields: fields{
				registry: New(client, WithPrefix("test")),
			},
			args: args{
				ctx:     context.Background(),
				service: testServer,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.fields.registry
			if err := r.Register(tt.args.ctx, tt.args.service); (err != nil) != tt.wantErr {
				t.Errorf("Register error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegistry_Deregister(t *testing.T) {
	testServer := &registry.ServiceInstance{
		ID:        "1",
		Name:      "test2",
		Version:   "v1.0.0",
		Endpoints: []string{"http://127.0.0.1:8080?isSecure=false"},
	}

	type args struct {
		ctx     context.Context
		service *registry.ServiceInstance
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		preFunc func(t *testing.T)
	}{
		{
			name: "normal",
			args: args{
				ctx:     context.Background(),
				service: testServer,
			},
			wantErr: false,
			preFunc: func(t *testing.T) {
				sc := []constant.ServerConfig{
					*constant.NewServerConfig("127.0.0.1", 8848),
				}

				cc := constant.ClientConfig{
					NamespaceId:         "public", // namespace id
					TimeoutMs:           5000,
					NotLoadCacheAtStart: true,
					LogDir:              "/tmp/nacos/log",
					CacheDir:            "/tmp/nacos/cache",
					LogLevel:            "debug",
				}

				// a more graceful way to create naming client
				client, err := clients.NewNamingClient(
					vo.NacosClientParam{
						ClientConfig:  &cc,
						ServerConfigs: sc,
					},
				)
				if err != nil {
					t.Fatal(err)
				}
				r := New(client)
				err = r.Register(context.Background(), testServer)
				if err != nil {
					t.Error(err)
				}
			},
		},
		{
			name: "error",
			args: args{
				ctx: context.Background(),
				service: &registry.ServiceInstance{
					ID:        "1",
					Name:      "test",
					Version:   "v1.0.0",
					Endpoints: []string{"127.0.0.1:8080"},
				},
			},
			wantErr: true,
		},
		{
			name: "errorPort",
			args: args{
				ctx: context.Background(),
				service: &registry.ServiceInstance{
					ID:        "1",
					Name:      "notExist",
					Version:   "v1.0.0",
					Endpoints: []string{"http://127.0.0.18080"},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := []constant.ServerConfig{
				*constant.NewServerConfig("127.0.0.1", 8848),
			}

			cc := constant.ClientConfig{
				NamespaceId:         "public", // namespace id
				TimeoutMs:           5000,
				NotLoadCacheAtStart: true,
				LogDir:              "/tmp/nacos/log",
				CacheDir:            "/tmp/nacos/cache",
				LogLevel:            "debug",
			}

			// a more graceful way to create naming client
			client, err := clients.NewNamingClient(
				vo.NacosClientParam{
					ClientConfig:  &cc,
					ServerConfigs: sc,
				},
			)
			if err != nil {
				t.Fatal(err)
			}
			r := New(client)
			if tt.preFunc != nil {
				tt.preFunc(t)
			}
			if err := r.Deregister(tt.args.ctx, tt.args.service); (err != nil) != tt.wantErr {
				t.Errorf("Deregister error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegistry_GetService(t *testing.T) {
	sc := []constant.ServerConfig{
		*constant.NewServerConfig("127.0.0.1", 8848),
	}

	cc := constant.ClientConfig{
		NamespaceId:         "public", // namespace id
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "debug",
	}

	// a more graceful way to create naming client
	client, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	r := New(client)
	testServer := &registry.ServiceInstance{
		ID:        "1",
		Name:      "test3",
		Version:   "v1.0.0",
		Endpoints: []string{"grpc://127.0.0.1:8080?isSecure=false"},
	}

	type fields struct {
		registry *Registry
	}
	type args struct {
		ctx         context.Context
		serviceName string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      []*registry.ServiceInstance
		wantErr   bool
		preFunc   func(t *testing.T)
		deferFunc func(t *testing.T)
	}{
		{
			name: "normal",
			preFunc: func(t *testing.T) {
				err = r.Register(context.Background(), testServer)
				if err != nil {
					t.Error(err)
				}
				time.Sleep(time.Second)
			},
			deferFunc: func(t *testing.T) {
				err = r.Deregister(context.Background(), testServer)
				if err != nil {
					t.Error(err)
				}
			},
			fields: fields{
				registry: r,
			},
			args: args{
				ctx:         context.Background(),
				serviceName: testServer.Name + "." + "grpc",
			},
			want: []*registry.ServiceInstance{{
				ID:        "127.0.0.1#8080#DEFAULT#DEFAULT_GROUP@@test3.grpc",
				Name:      "DEFAULT_GROUP@@test3.grpc",
				Version:   "v1.0.0",
				Metadata:  map[string]string{"version": "v1.0.0", "kind": "grpc"},
				Endpoints: []string{"grpc://127.0.0.1:8080"},
			}},
			wantErr: false,
		},
		{
			name: "errorNotExist",
			fields: fields{
				registry: r,
			},
			args: args{
				ctx:         context.Background(),
				serviceName: "notExist",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preFunc != nil {
				tt.preFunc(t)
			}
			if tt.deferFunc != nil {
				defer tt.deferFunc(t)
			}
			r := tt.fields.registry
			got, err := r.GetService(tt.args.ctx, tt.args.serviceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetService error = %v, wantErr %v", err, tt.wantErr)
				t.Errorf("GetService got = %v", got)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetService got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_Watch(t *testing.T) {
	sc := []constant.ServerConfig{
		*constant.NewServerConfig("127.0.0.1", 8848),
	}

	cc := constant.ClientConfig{
		NamespaceId:         "public", // namespace id
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "debug",
	}

	// a more graceful way to create naming client
	client, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	r := New(client)

	testServer := &registry.ServiceInstance{
		ID:        "1",
		Name:      "test4",
		Version:   "v1.0.0",
		Endpoints: []string{"grpc://127.0.0.1:8080?isSecure=false"},
	}

	cancelCtx, cancel := context.WithCancel(context.Background())
	type fields struct {
		registry *Registry
	}
	type args struct {
		ctx         context.Context
		serviceName string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		want        []*registry.ServiceInstance
		processFunc func(t *testing.T)
	}{
		{
			name: "normal",
			fields: fields{
				registry: New(client),
			},
			args: args{
				ctx:         context.Background(),
				serviceName: testServer.Name + "." + "grpc",
			},
			wantErr: false,
			want: []*registry.ServiceInstance{{
				ID:        "127.0.0.1#8080#DEFAULT#DEFAULT_GROUP@@test4.grpc",
				Name:      "DEFAULT_GROUP@@test4.grpc",
				Version:   "v1.0.0",
				Metadata:  map[string]string{"version": "v1.0.0", "kind": "grpc"},
				Endpoints: []string{"grpc://127.0.0.1:8080"},
			}},
			processFunc: func(t *testing.T) {
				err = r.Register(context.Background(), testServer)
				if err != nil {
					t.Error(err)
				}
			},
		},
		{
			name: "ctxCancel",
			fields: fields{
				registry: r,
			},
			args: args{
				ctx:         cancelCtx,
				serviceName: testServer.Name,
			},
			wantErr: true,
			want:    nil,
			processFunc: func(t *testing.T) {
				cancel()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.fields.registry
			watch, err := r.Watch(tt.args.ctx, tt.args.serviceName)
			if err != nil {
				t.Error(err)
				return
			}
			defer func() {
				err = watch.Stop()
				if err != nil {
					t.Error(err)
				}
			}()
			_, err = watch.Next()
			if err != nil {
				t.Error(err)
				return
			}

			if tt.processFunc != nil {
				tt.processFunc(t)
			}

			want, err := watch.Next()
			if (err != nil) != tt.wantErr {
				t.Errorf("Watch error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(want, tt.want) {
				t.Errorf("Watch watcher = %v, want %v", watch, tt.want)
			}
		})
	}
}
