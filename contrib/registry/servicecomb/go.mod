module github.com/go-kratos/kratos/contrib/registry/servicecomb/v2

go 1.21

require (
	github.com/go-chassis/cari v0.6.0
	github.com/go-chassis/sc-client v0.6.1-0.20210615014358-a45e9090c751
	github.com/go-kratos/kratos/v2 v2.8.3
	github.com/gofrs/uuid v4.2.0+incompatible
)

require (
	github.com/cenkalti/backoff v2.0.0+incompatible // indirect
	github.com/go-chassis/foundation v0.4.0 // indirect
	github.com/go-chassis/openlog v1.1.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/gorilla/websocket v1.4.3-0.20210424162022-e8629af678b7 // indirect
	golang.org/x/net v0.23.0 // indirect
)

replace github.com/go-kratos/kratos/v2 => ../../../
