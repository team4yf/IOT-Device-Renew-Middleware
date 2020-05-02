// the core grpc server

package main

import (
	"fmt"
	pb "github/team4yf/IOT-Device-Renew-Middleware/drm"
	"log"
	"net"
	"strconv"
	"strings"

	"time"

	"github.com/go-redis/redis/v7"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	PORT         = "5009"
	REDIS_HOST   = "localhost"
	REDIS_PORT   = "6379"
	REDIS_DB     = 13
	REDIS_PASS   = "admin123"
	REDIS_PREFIX = "drm"
	MQTT_URL     = "www.ruichen.top:1883"
	MQTT_USER    = "admin"
	MQTT_PASS    = "123123123"
)

var (
	client     *redis.Client
	mqttClient mqtt.Client
)

func mqttConn() mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", MQTT_URL))
	opts.SetUsername(MQTT_USER)
	opts.SetPassword(MQTT_PASS)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Fatal(err)
	}
	return client
}

func init() {
	client = redis.NewClient(&redis.Options{
		Addr:     REDIS_HOST + ":" + REDIS_PORT,
		Password: REDIS_PASS,
		DB:       REDIS_DB,
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)

	mqttClient = mqttConn()
}

type server struct{}

func main() {
	go subscribe()
	// serve & bind a grpc channel
	log.Printf("startup")
	lis, err := net.Listen("tcp", ":"+PORT)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterDeviceRenewServer(s, &server{})
	s.Serve(lis)
	log.Println("grpc serve in :%s", PORT)

}

func subscribe() {
	pubsub := client.Subscribe(fmt.Sprintf("__keyevent@%d__:expired", REDIS_DB))
	defer pubsub.Close()
	for {
		msg, _ := pubsub.ReceiveMessage()
		deviceKey := msg.Payload
		// filter the other key event
		if !strings.HasPrefix(deviceKey, REDIS_PREFIX) {
			continue
		}

		log.Printf("expired " + deviceKey)
		// publish the event
		mqttClient.Publish("$drm/offline", 0, false, deviceKey)
	}
}

func (s *server) Renew(ctx context.Context, request *pb.RenewRequest) (response *pb.RenewResponse, err error) {
	device := fmt.Sprintf("%s:%s:%s", REDIS_PREFIX, request.Project, request.Device)
	expire := request.Expire
	isOk, err := renew(device, expire)
	if err != nil {
		return nil, err
	}
	response = &pb.RenewResponse{
		IsOk: isOk,
	}
	return response, nil
}

func (s *server) Check(ctx context.Context, request *pb.CheckRequest) (response *pb.CheckResponse, err error) {
	device := fmt.Sprintf("%s:%s:%s", REDIS_PREFIX, request.Project, request.Device)
	isOk, err := check(device)
	response = &pb.CheckResponse{
		IsOk: false,
	}
	if err != nil {
		return nil, err
	}
	response.IsOk = isOk
	return response, nil
}

// renew the device
func renew(device string, ex int64) (bool, error) {
	err := client.Set(device, strconv.FormatInt(time.Now().Unix(), 10), time.Duration(ex)*time.Second).Err()
	if err != nil {
		return false, err
	}
	return true, nil
}

// check the device
func check(device string) (bool, error) {
	cmd := client.Get(device)
	if cmd.Err() != nil {
		return false, cmd.Err()
	}
	log.Printf("%s=>val: %s", device, cmd.Val())
	return true, nil
}
