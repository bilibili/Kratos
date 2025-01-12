module github.com/go-kratos/kratos/contrib/log/logrus/v2

go 1.21

require (
	github.com/go-kratos/kratos/v2 v2.8.3
	github.com/sirupsen/logrus v1.8.1
)

require golang.org/x/sys v0.18.0 // indirect

replace github.com/go-kratos/kratos/v2 => ../../../
