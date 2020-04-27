package main

import (
	"log"

	pb "github/team4yf/IOT-Device-Renew-Middleware/drm"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	networkType = "tcp"
	address     = "localhost:5009"
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
	Check(client)

}

// implement the interface
func Renew(client pb.DeviceRenewClient) {
	var request pb.RenewRequest
	request.Project = "testa"
	request.Device = "/dd/aa"
	request.Expire = 1

	response, _ := client.Renew(context.Background(), &request) //调用远程方法

	log.Printf("Renew response result  %#v", response.IsOk)
}

func Check(client pb.DeviceRenewClient) {
	var request pb.CheckRequest
	request.Project = "testa"
	request.Device = "/dd/aa"

	response, _ := client.Check(context.Background(), &request) //调用远程方法

	log.Printf("Check response result  %#v", response.IsOk)
}
