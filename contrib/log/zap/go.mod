module github.com/go-kratos/kratos/contrib/log/zap/v2

go 1.21

require (
	github.com/go-kratos/kratos/v2 v2.8.3
	go.uber.org/zap v1.26.0
)

require go.uber.org/multierr v1.11.0 // indirect

replace github.com/go-kratos/kratos/v2 => ../../../
