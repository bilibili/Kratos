package nacos

import (
	"context"
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"

	"github.com/go-kratos/kratos/v2/registry"
)

var _ registry.Watcher = (*watcher)(nil)

type watcher struct {
	serviceName string
	clusters    []string
	groupName   string
	ctx         context.Context
	cancel      context.CancelFunc
	watchChan   chan struct{}
	cli         naming_client.INamingClient
	kind        string
}

func newWatcher(ctx context.Context, cli naming_client.INamingClient, serviceName, groupName, kind string, clusters []string) (*watcher, error) {
	w := &watcher{
		serviceName: serviceName,
		clusters:    clusters,
		groupName:   groupName,
		cli:         cli,
		kind:        kind,
		watchChan:   make(chan struct{}, 1),
	}
	w.ctx, w.cancel = context.WithCancel(ctx)

	e := w.cli.Subscribe(&vo.SubscribeParam{
		ServiceName: serviceName,
		Clusters:    clusters,
		GroupName:   groupName,
		SubscribeCallback: func(services []model.Instance, err error) {
			w.watchChan <- struct{}{}
		},
	})
	return w, e
}

func (w *watcher) Next() ([]*registry.ServiceInstance, error) {
	select {
	case <-w.ctx.Done():
		return nil, w.ctx.Err()
	case <-w.watchChan:
	}
	res, err := w.cli.GetService(vo.GetServiceParam{
		ServiceName: w.serviceName,
		GroupName:   w.groupName,
		Clusters:    w.clusters,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*registry.ServiceInstance, 0, len(res.Hosts))
	for _, in := range res.Hosts {
		kind := w.kind
		if k, ok := in.Metadata["kind"]; ok {
			kind = k
		}
		items = append(items, &registry.ServiceInstance{
			ID:        in.InstanceId,
			Name:      res.Name,
			Version:   in.Metadata["version"],
			Metadata:  in.Metadata,
			Endpoints: []string{fmt.Sprintf("%s://%s:%d", kind, in.Ip, in.Port)},
		})
	}
	return items, nil
}

func (w *watcher) Stop() error {
	w.cancel()
	return w.cli.Unsubscribe(&vo.SubscribeParam{
		ServiceName: w.serviceName,
		GroupName:   w.groupName,
		Clusters:    w.clusters,
	})
}
