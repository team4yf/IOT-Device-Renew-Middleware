package main

import (
	"log"
	"time"

	pb "github/team4yf/IOT-Device-Renew-Middleware/drm"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	networkType = "tcp"
	address     = "localhost:5009"
	PROJ        = "foo"
	DEVICE      = "bar"
)

func main() {
	//建立连接
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewDeviceRenewClient(conn)
	Renew(client)
	time.Sleep(time.Second)
	Check(client)

}

// implement the interface
func Renew(client pb.DeviceRenewClient) {
	var request pb.RenewRequest
	request.Project = PROJ
	request.Device = DEVICE
	request.Expire = 30

	response, err := client.Renew(context.Background(), &request) //调用远程方法
	if err != nil {
		log.Printf("Renew response error  %#v", err)
	} else {
		log.Printf("Renew response result  %#v", response.IsOk)
	}
}

func Check(client pb.DeviceRenewClient) {
	var request pb.CheckRequest
	request.Project = PROJ
	request.Device = DEVICE

	response, err := client.Check(context.Background(), &request) //调用远程方法
	if err != nil {
		log.Printf("Check response error  %#v", err)
	} else {
		log.Printf("Check response result  %#v", response.IsOk)
	}

}
