package validate

import (
	"context"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/go-kratos/kratos/contrib/middleware/validate/v2/internal/testdata"
	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
)

type testcase struct {
	name string
	req  proto.Message
	err  bool
}

func TestTable(t *testing.T) {
	var mock middleware.Handler = func(context.Context, interface{}) (interface{}, error) { return nil, nil }

	tests := []testcase{
		{
			name: "valid_legacy",
			req:  &testdata.Legacy{Name: "testcase", Age: 19},
			err:  false,
		},
		{
			name: "invalid_legacy1",
			req:  &testdata.Legacy{Name: "testcase", Age: 10},
			err:  true,
		},
		{
			name: "invalid_legacy2",
			req:  &testdata.Legacy{Name: "test", Age: 100},
			err:  true,
		},
		{
			name: "valid_mixed",
			req:  &testdata.Mixed{Name: "testcase", Age: 19},
			err:  false,
		},
		{
			name: "invalid_mixed1",
			req:  &testdata.Mixed{Name: "testcase", Age: 10},
			err:  true,
		},
		{
			name: "invalid_mixed2",
			req:  &testdata.Mixed{Name: "test", Age: 100},
			err:  true,
		},
		{
			name: "valid_modern",
			req:  &testdata.Modern{Name: "testcase", Age: 19},
			err:  false,
		},
		{
			name: "invalid_modern1",
			req:  &testdata.Modern{Name: "testcase", Age: 10},
			err:  true,
		},
		{
			name: "invalid_modern2",
			req:  &testdata.Modern{Name: "test", Age: 100},
			err:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handle := ProtoValidate()(mock)
			_, err := handle(context.Background(), test.req)
			expect := test.err
			actual := kerrors.IsBadRequest(err)
			if expect != actual {
				t.Errorf("case %s expect %v, actual %v, err %v", test.name, expect, actual, err)
			}
		})
	}
}
