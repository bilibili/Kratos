package jwt

import (
	"context"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type headerCarrier http.Header

func (hc headerCarrier) Get(key string) string { return http.Header(hc).Get(key) }

func (hc headerCarrier) Set(key string, value string) { http.Header(hc).Set(key, value) }

// Keys lists the keys stored in this carrier.
func (hc headerCarrier) Keys() []string {
	keys := make([]string, 0, len(hc))
	for k := range http.Header(hc) {
		keys = append(keys, k)
	}
	return keys
}

func newTokenHeader(token string) *headerCarrier {
	header := &headerCarrier{}
	header.Set(JWTHeaderKey, token)
	return header
}

type Transport struct {
	kind      transport.Kind
	endpoint  string
	operation string
	reqHeader transport.Header
}

func (tr *Transport) Kind() transport.Kind {
	return tr.kind
}
func (tr *Transport) Endpoint() string {
	return tr.endpoint
}
func (tr *Transport) Operation() string {
	return tr.operation
}
func (tr *Transport) RequestHeader() transport.Header {
	return tr.reqHeader
}
func (tr *Transport) ReplyHeader() transport.Header {
	return nil
}

func TestJwtToken(t *testing.T) {
	var testKey = "testKey"
	mapClaims := jwt.MapClaims{}
	mapClaims["name"] = "xiaoli"
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, mapClaims)
	token, err := claims.SignedString([]byte(testKey))
	if err != nil {
		panic(err)
	}
	// todo add test case
	tests := []struct {
		name          string
		ctx           context.Context
		signingMethod jwt.SigningMethod
	}{
		{
			name:          "normal",
			ctx:           transport.NewServerContext(context.Background(), &Transport{reqHeader: newTokenHeader(token)}),
			signingMethod: jwt.SigningMethodHS256,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var testToken interface{}
			next := func(ctx context.Context, req interface{}) (interface{}, error) {
				t.Log(req)
				testToken = ctx.Value(JWTClaimsContextKey)
				return "reply", nil
			}
			server := Server(testKey, test.signingMethod)(next)
			_, err2 := server(test.ctx, test.name)
			assert.Nil(t, err2)
			assert.NotNil(t, testToken)
		})
	}

}
