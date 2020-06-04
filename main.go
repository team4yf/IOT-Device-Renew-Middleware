// the core grpc server

package main

import (
	"fmt"
	pb "github/team4yf/IOT-Device-Renew-Middleware/drm"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var (
	PORT              = "3009"
	REDIS_HOST        = "localhost"
	REDIS_PORT        = "6379"
	REDIS_DB          = 13
	REDIS_PASS        = "admin123"
	REDIS_PREFIX      = "drm"
	MQTT_URL          = "www.ruichen.top:1883"
	MQTT_USER         = "admin"
	MQTT_PASS         = "123123123"
	MQTT_EVENT_PREFIX = "^drm"
	MQTT_EVENT_QOS    = 0
	MQTT_EVENT_RETAIN = false
)

var (
	client     *redis.Client
	mqttClient MQTT.Client
)

func mqttConn() MQTT.Client {
	opts := MQTT.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s", MQTT_URL))
	// opts.SetClientID("go-simple")
	opts.SetUsername(MQTT_USER)
	opts.SetPassword(MQTT_PASS)
	client := MQTT.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error().Error())
	}
	return client
}

func initGetEnv() {

	if "" != os.Getenv("SERVE_PORT") {
		PORT = os.Getenv("SERVE_PORT")
	}
	if "" != os.Getenv("REDIS_HOST") {
		REDIS_HOST = os.Getenv("REDIS_HOST")
	}
	if "" != os.Getenv("REDIS_PORT") {
		REDIS_PORT = os.Getenv("REDIS_PORT")
	}
	if "" != os.Getenv("REDIS_PASS") {
		REDIS_PASS = os.Getenv("REDIS_PASS")
	}
	if "" != os.Getenv("REDIS_DB") {
		REDIS_DB, _ = strconv.Atoi(os.Getenv("REDIS_DB"))
	}
	if "" != os.Getenv("REDIS_PREFIX") {
		REDIS_PREFIX = os.Getenv("REDIS_PREFIX")
	}
	if "" != os.Getenv("MQTT_URL") {
		MQTT_URL = os.Getenv("MQTT_URL")
	}
	if "" != os.Getenv("MQTT_USER") {
		MQTT_USER = os.Getenv("MQTT_USER")
	}
	if "" != os.Getenv("MQTT_PASS") {
		MQTT_PASS = os.Getenv("MQTT_PASS")
	}
	if "" != os.Getenv("MQTT_EVENT_PREFIX") {
		MQTT_EVENT_PREFIX = os.Getenv("MQTT_EVENT_PREFIX")
	}
	if "" != os.Getenv("MQTT_EVENT_QOS") {
		MQTT_EVENT_QOS, _ = strconv.Atoi(os.Getenv("MQTT_EVENT_QOS"))
	}
	if "" != os.Getenv("MQTT_EVENT_RETAIN") {
		MQTT_EVENT_RETAIN, _ = strconv.ParseBool(os.Getenv("MQTT_EVENT_RETAIN"))
	}
}

func init() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	initGetEnv()
	client = redis.NewClient(&redis.Options{
		Addr:     REDIS_HOST + ":" + REDIS_PORT,
		Password: REDIS_PASS,
		DB:       REDIS_DB,
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Fatal("redis cant connect ", err)
	}

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
	log.Printf("grpc serve in :%s\n", PORT)
	s.Serve(lis)

}

func subscribe() {
	pubsub := client.Subscribe(fmt.Sprintf("__keyevent@%d__:expired", REDIS_DB))
	defer pubsub.Close()
	// log.Printf("Redis expired key event subscribe success!\n")
	for {
		msg, _ := pubsub.ReceiveMessage()
		deviceKey := msg.Payload
		// filter the other key event
		if !strings.HasPrefix(deviceKey, REDIS_PREFIX) {
			continue
		}

		isOk, err := publishOfflineEvent(deviceKey)
		if !isOk {
			log.Println("publishOfflineEvent failed, error: ", err)
		}

	}
}

func (s *server) Renew(ctx context.Context, request *pb.RenewRequest) (response *pb.RenewResponse, err error) {
	device := fmt.Sprintf("%s:%s:%s", REDIS_PREFIX, request.Project, request.Device)

	// if the device not exist, publish a online event
	isOk, err := check(device)
	if err != nil {
		return nil, err
	}
	if !isOk {
		go func() {
			pushed, _ := publishOnlineEvent(device)
			log.Println("pushed: ", pushed)
		}()
	}
	expire := request.Expire
	isOk, err = renew(device, expire)
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

func publishOnlineEvent(deviceKey string) (bool, error) {
	proj, deviceID := splitDeviceKey(deviceKey)
	// publish the event
	token := mqttClient.Publish(MQTT_EVENT_PREFIX+"/online/"+proj, byte(MQTT_EVENT_QOS), MQTT_EVENT_RETAIN, deviceID)
	token.Wait()
	return true, nil
}

func publishOfflineEvent(deviceKey string) (bool, error) {
	proj, deviceID := splitDeviceKey(deviceKey)
	// publish the event
	token := mqttClient.Publish(MQTT_EVENT_PREFIX+"/offline/"+proj, byte(MQTT_EVENT_QOS), MQTT_EVENT_RETAIN, deviceID)
	token.Wait()
	return true, nil
}

// split the redis key,
// like: drm:foo:bar
// it should return foo, bar
func splitDeviceKey(deviceKey string) (string, string) {
	subStrs := strings.Split(deviceKey, ":")
	proj := subStrs[1]
	deviceID := strings.Join(subStrs[1:], ":")
	return proj, deviceID
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
	cmd := client.Exists(device)
	if cmd.Err() != nil {
		return false, cmd.Err()
	}
	log.Printf(device, "=>val: ", cmd.Val())
	return cmd.Val() == 1, nil
}
