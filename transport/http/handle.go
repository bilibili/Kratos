package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/go-kratos/kratos/v2/encoding"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport/http/binding"
)

const (
	// SupportPackageIsVersion1 These constants should not be referenced from any other code.
	SupportPackageIsVersion1 = true

	baseContentType = "application"
)

var (
	acceptHeader      = http.CanonicalHeaderKey("Accept")
	contentTypeHeader = http.CanonicalHeaderKey("Content-Type")
)

// DecodeRequestFunc is decode request func.
type DecodeRequestFunc func(*http.Request, interface{}) error

// EncodeResponseFunc is encode response func.
type EncodeResponseFunc func(http.ResponseWriter, *http.Request, interface{}) error

// EncodeErrorFunc is encode error func.
type EncodeErrorFunc func(http.ResponseWriter, *http.Request, error)

// StatusCoder is returns the HTTPStatus code.
type StatusCoder interface {
	HTTPStatus() int
}

// HandleOption is handle option.
type HandleOption func(*HandleOptions)

// HandleOptions is handle options.
type HandleOptions struct {
	Decode     DecodeRequestFunc
	Encode     EncodeResponseFunc
	Error      EncodeErrorFunc
	Middleware middleware.Middleware
}

// DefaultHandleOptions returns a default handle options.
func DefaultHandleOptions() HandleOptions {
	return HandleOptions{
		Decode: DecodeRequest,
		Encode: EncodeResponse,
		Error:  EncodeError,
	}
}

// RequestDecoder with request decoder.
func RequestDecoder(dec DecodeRequestFunc) HandleOption {
	return func(o *HandleOptions) {
		o.Decode = dec
	}
}

// ResponseEncoder with response encoder.
func ResponseEncoder(en EncodeResponseFunc) HandleOption {
	return func(o *HandleOptions) {
		o.Encode = en
	}
}

// ErrorEncoder with error encoder.
func ErrorEncoder(en EncodeErrorFunc) HandleOption {
	return func(o *HandleOptions) {
		o.Error = en
	}
}

// Middleware with middleware option.
func Middleware(m middleware.Middleware) HandleOption {
	return func(o *HandleOptions) {
		o.Middleware = m
	}
}

// DecodeRequest decodes the request body to object.
func DecodeRequest(req *http.Request, v interface{}) error {
	subtype := contentSubtype(req.Header.Get(contentTypeHeader))
	if codec := encoding.GetCodec(subtype); codec != nil {
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return err
		}
		return codec.Unmarshal(data, v)
	}
	return binding.BindForm(req, v)
}

// EncodeResponse encodes the object to the HTTP response.
func EncodeResponse(w http.ResponseWriter, r *http.Request, v interface{}) error {
	for _, accept := range r.Header[acceptHeader] {
		if codec := encoding.GetCodec(contentSubtype(accept)); codec != nil {
			data, err := codec.Marshal(v)
			if err != nil {
				return err
			}
			w.Header().Set(contentTypeHeader, contentType(codec.Name()))
			w.Write(data)
			return nil
		}
	}
	return json.NewEncoder(w).Encode(v)
}

// EncodeError encodes the erorr to the HTTP response.
func EncodeError(w http.ResponseWriter, r *http.Request, err error) {
	se, ok := errors.FromError(err)
	if !ok {
		se = &errors.StatusError{
			Code:    2,
			Reason:  "",
			Message: err.Error(),
		}
	}
	w.WriteHeader(se.HTTPStatus())
	EncodeResponse(w, r, se)
}

func contentType(subtype string) string {
	return strings.Join([]string{baseContentType, subtype}, "/")
}

func contentSubtype(contentType string) string {
	if contentType == baseContentType {
		return ""
	}
	if !strings.HasPrefix(contentType, baseContentType) {
		return ""
	}
	switch contentType[len(baseContentType)] {
	case '/', ';':
		if i := strings.Index(contentType, ";"); i != -1 {
			return contentType[len(baseContentType)+1 : i]
		}
		return contentType[len(baseContentType)+1:]
	default:
		return ""
	}
}
