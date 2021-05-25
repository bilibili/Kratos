// Code generated by protoc-gen-go-http. DO NOT EDIT.

package testproto

import (
	context "context"
	http1 "github.com/go-kratos/kratos/v2/transport/http"
	binding "github.com/go-kratos/kratos/v2/transport/http/binding"
	mux "github.com/gorilla/mux"
	http "net/http"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the kratos package it is being compiled against.
var _ = new(http.Request)
var _ = new(context.Context)
var _ = binding.MapProto
var _ = mux.NewRouter

const _ = http1.SupportPackageIsVersion1

type EchoServiceHandler interface {
	Echo(context.Context, *SimpleMessage) (*SimpleMessage, error)

	EchoBody(context.Context, *SimpleMessage) (*SimpleMessage, error)

	EchoDelete(context.Context, *SimpleMessage) (*SimpleMessage, error)

	EchoPatch(context.Context, *DynamicMessageUpdate) (*DynamicMessageUpdate, error)

	EchoResponseBody(context.Context, *DynamicMessageUpdate) (*DynamicMessageUpdate, error)
}

func NewEchoServiceHandler(srv EchoServiceHandler, opts ...http1.HandleOption) http.Handler {
	r := mux.NewRouter()

	r.Handle("/v1/example/echo/{id}/{num}", http1.NewHandler(srv.Echo, opts...)).Methods("GET")

	r.Handle("/v1/example/echo/{id}/{num}/{lang}", http1.NewHandler(srv.Echo, opts...)).Methods("GET")

	r.Handle("/v1/example/echo1/{id}/{line_num}/{status.note}", http1.NewHandler(srv.Echo, opts...)).Methods("GET")

	r.Handle("/v1/example/echo2/{no.note}", http1.NewHandler(srv.Echo, opts...)).Methods("GET")

	r.Handle("/v1/example/echo/{id}", http1.NewHandler(srv.Echo, opts...)).Methods("POST")

	r.Handle("/v1/example/echo_body", http1.NewHandler(srv.EchoBody, opts...)).Methods("POST")

	r.Handle("/v1/example/echo_response_body", http1.NewHandler(srv.EchoResponseBody, opts...)).Methods("POST")

	r.Handle("/v1/example/echo_delete/{id}/{num}", http1.NewHandler(srv.EchoDelete, opts...)).Methods("DELETE")

	r.Handle("/v1/example/echo_patch", http1.NewHandler(srv.EchoPatch, opts...)).Methods("PATCH")

	return r
}

type EchoServiceHttpClient interface {
	Echo(ctx context.Context, req *SimpleMessage, opts ...http1.CallOption) (rsp *SimpleMessage, err error)

	EchoBody(ctx context.Context, req *SimpleMessage, opts ...http1.CallOption) (rsp *SimpleMessage, err error)

	EchoDelete(ctx context.Context, req *SimpleMessage, opts ...http1.CallOption) (rsp *SimpleMessage, err error)

	EchoPatch(ctx context.Context, req *DynamicMessageUpdate, opts ...http1.CallOption) (rsp *DynamicMessageUpdate, err error)

	EchoResponseBody(ctx context.Context, req *DynamicMessageUpdate, opts ...http1.CallOption) (rsp *DynamicMessageUpdate, err error)
}

type EchoServiceHttpClientImpl struct {
	cc *http1.Client
}

func NewEchoServiceHttpClient(client *http1.Client) EchoServiceHttpClient {
	return &EchoServiceHttpClientImpl{client}
}

func (c *EchoServiceHttpClientImpl) Echo(ctx context.Context, in *SimpleMessage, opts ...http1.CallOption) (out *SimpleMessage, err error) {
	path := binding.ProtoPath("/v1/example/echo/{id}", in)
	out = &SimpleMessage{}

	err = c.cc.Invoke(ctx, path, nil, &out, http1.Method("POST"), http1.PathPattern("/v1/example/echo/{id}"))

	if err != nil {
		return
	}
	return
}

func (c *EchoServiceHttpClientImpl) EchoBody(ctx context.Context, in *SimpleMessage, opts ...http1.CallOption) (out *SimpleMessage, err error) {
	path := binding.ProtoPath("/v1/example/echo_body", in)
	out = &SimpleMessage{}

	err = c.cc.Invoke(ctx, path, in, &out, http1.Method("POST"), http1.PathPattern("/v1/example/echo_body"))

	if err != nil {
		return
	}
	return
}

func (c *EchoServiceHttpClientImpl) EchoDelete(ctx context.Context, in *SimpleMessage, opts ...http1.CallOption) (out *SimpleMessage, err error) {
	path := binding.ProtoPath("/v1/example/echo_delete/{id}/{num}", in)
	out = &SimpleMessage{}

	err = c.cc.Invoke(ctx, path, nil, &out, http1.Method("DELETE"), http1.PathPattern("/v1/example/echo_delete/{id}/{num}"))

	if err != nil {
		return
	}
	return
}

func (c *EchoServiceHttpClientImpl) EchoPatch(ctx context.Context, in *DynamicMessageUpdate, opts ...http1.CallOption) (out *DynamicMessageUpdate, err error) {
	path := binding.ProtoPath("/v1/example/echo_patch", in)
	out = &DynamicMessageUpdate{}

	err = c.cc.Invoke(ctx, path, in.Body, &out, http1.Method("PATCH"), http1.PathPattern("/v1/example/echo_patch"))

	if err != nil {
		return
	}
	return
}

func (c *EchoServiceHttpClientImpl) EchoResponseBody(ctx context.Context, in *DynamicMessageUpdate, opts ...http1.CallOption) (out *DynamicMessageUpdate, err error) {
	path := binding.ProtoPath("/v1/example/echo_response_body", in)
	out = &DynamicMessageUpdate{}

	err = c.cc.Invoke(ctx, path, in, &out.Body, http1.Method("POST"), http1.PathPattern("/v1/example/echo_response_body"))

	if err != nil {
		return
	}
	return
}
