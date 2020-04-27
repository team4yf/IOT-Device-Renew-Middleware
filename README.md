# IOT-Device-Renew-Middleware
IOT Device Renew Middleware


[Google Doc](https://docs.google.com/document/d/1kdXLDb_kQuah-iinMXNIq_qNe8Nq0Pyg59AvkRdE6hI/edit#heading=h.mce78li9b3g5)

Project Code: [DRM].

It’s a middleware for iot devices to renew their online status.

2 important logic: renew, timeup

Renew
1. Device renews itself 10s.
IOT -> DRM -> renew 10s 
2. DRM publishes an event notification right now if the device does not exist.
	DRM -> notify -> App

Timeup
1. DRM will publish an event notification in 10s if the device does not renew again.
DRM -> notify -> App
2. Check device
	App -> DRM -> check -> T/F


We supply 2 gRPC API
  - `/renew()`
  - `/check()`
There are 2 event notifications
  - `$drm/online/:device`
  - `$drm/offline/:device`


#### go for gRPC

go get -u google.golang.org/grpc
go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
go get -u golang.org/x/net
protoc --go_out=plugins=grpc:. *.proto

export PATH="$PATH:$(go env GOPATH)/bin"
export GOPATH="/home/wangfan/Workspace/golang/"