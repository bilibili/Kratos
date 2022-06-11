package main

import (
	"bytes"
	"strings"
	"text/template"
)

var clientTemplate = `{{range .ClientInfoList }}
type {{ .ServiceName }}GRPCClient struct {
	cli {{ .ServiceName }}Client
}

//New{{ .ServiceName }}GRPCClient create grpc client for kratos
func New{{ .ServiceName }}GRPCClient(opts ...grpc.ClientOption) (cli *{{ .ServiceName }}GRPCClient, err error) {
	{{ if .Endpoint }}
	conn, ok := connMap["{{ .Endpoint }}"]
	if !ok {
		opts = append(opts, grpc.WithEndpoint("{{ .Endpoint }}"))
		conn, err = grpc.DialInsecure(context.Background(), opts...)
		if err != nil {
			return nil, err
		}
		connMap["{{ .Endpoint }}"] = conn
	}
	{{- else }}
	conn, err := grpc.DialInsecure(context.Background(), opts...)
	{{- end }}
	if err != nil {
		return nil, err
	}
	client := New{{ .ServiceName }}Client(conn)
	return &{{ .ServiceName }}GRPCClient{cli:client}, nil
}
{{- end }}
`

type ClientInfo struct {
	ServiceName string // proto service
	Endpoint    string // default_host
}

type ClientTemplate struct {
	ClientInfoList []ClientInfo
}

//NewClientTemplate new client template
func NewClientTemplate() *ClientTemplate {
	return &ClientTemplate{
		ClientInfoList: make([]ClientInfo, 0, 5),
	}
}

func (receiver *ClientTemplate) AppendClientInfo(serviceName, endpoint string) {
	receiver.ClientInfoList = append(receiver.ClientInfoList, ClientInfo{
		ServiceName: serviceName,
		Endpoint:    endpoint,
	})
}

//Parse create content
func (receiver *ClientTemplate) execute() string {
	parser, err := template.New("clientTemplate").Parse(clientTemplate)
	if err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	if err := parser.Execute(buf, receiver); err != nil {
		panic(err)
	}
	return strings.Trim(buf.String(), "\r\n")
}
