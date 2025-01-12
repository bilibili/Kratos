package validate

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"

	"github.com/bufbuild/protovalidate-go"
	"github.com/bufbuild/protovalidate-go/legacy"
	"google.golang.org/protobuf/proto"
)

// ProtoValidate is a middleware that validates the request message with [protovalidate](https://github.com/bufbuild/protovalidate)
func ProtoValidate() middleware.Middleware {
	validator, err := protovalidate.New(
		// Some projects may still be using PGV, turn on legacy support to handle this.
		legacy.WithLegacySupport(legacy.ModeMerge),
	)
	if err != nil {
		panic(fmt.Errorf("protovalidate.New() error: %w", err))
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if msg, ok := req.(proto.Message); ok {
				if err := validator.Validate(msg); err != nil {
					return nil, errors.BadRequest("VALIDATOR", err.Error()).WithCause(err)
				}
			}
			return handler(ctx, req)
		}
	}
}
