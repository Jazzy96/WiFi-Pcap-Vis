module wifi-pcap-demo/router_agent

go 1.20 // Updated to support newer protobuf/grpc features

require (
	google.golang.org/grpc v1.64.0 // Updated to match generated code requirements
	google.golang.org/protobuf v1.33.0 // Updated to a more recent version, consider running 'go get -u google.golang.org/protobuf' and 'go mod tidy'
)

require (
	github.com/golang/protobuf v1.5.4 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
)
