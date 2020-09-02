//Pacakge main the main entry

package main

// import fpm-server core & mqtt-client plugin & redis plugin
import (
	"encoding/json"
	"fmt"
	"github/team4yf/IOT-Device-Renew-Middleware/message"
	"log"
	"strconv"
	"strings"
	"time"

	_ "github.com/team4yf/fpm-go-plugin-cache-redis/plugin"
	_ "github.com/team4yf/fpm-go-plugin-mqtt-client/plugin"
	"github.com/team4yf/yf-fpm-server-go/fpm"
)

const (
	redisPrefix = "DRM"
)

func main() {

	fpmApp := fpm.New()

	fpmApp.Init()
	redisDB := fpmApp.GetConfig("redis.db").(float64)
	redisBizPrefix := fpmApp.GetConfig("redis.prefix").(string)
	//sub topics
	fpmApp.Execute("redis.subscribe", &fpm.BizParam{
		"topic": fmt.Sprintf("__keyevent@%d__:expired", int(redisDB)),
	})
	//执行订阅的函数
	fpmApp.Execute("mqttclient.subscribe", &fpm.BizParam{
		"topics": []string{"$drm/+/renew", "$drm/+/message", "$d2s/+/+/feedback"},
	})
	//catch the message
	fpmApp.Subscribe("#redis/receive", func(_ string, data interface{}) {
		//data 通常是 byte[] 类型，可以转成 string 或者 map
		body := data.(map[string]interface{})
		t := body["topic"].(string)
		p := body["payload"].(string)
		fpmApp.Logger.Debugf("redis topic: %s, payload: %+v", t, p)
		//get the data
		deviceKey := p

		// filter the other key event
		if !strings.HasPrefix(deviceKey, redisBizPrefix+":"+redisPrefix) {
			return
		}

		isOk, err := publishOfflineEvent(deviceKey)
		if !isOk {
			fpmApp.Logger.Infof("publishOfflineEvent failed, error: %v", err)
		}

	})
	fpmApp.Subscribe("#mqtt/receive", func(_ string, data interface{}) {
		body := data.(map[string]interface{})
		fpmApp.Logger.Debugf("mqtt data: %+v", body)

		t := body["topic"].(string)
		p := body["payload"].([]byte)
		switch {
		//https://shimo.im/docs/bJaoNiMc4yEfkRSt#anchor-9DXU
		case strings.HasSuffix(t, "renew"):
			msg := message.RenewMessage{}
			if err := json.Unmarshal(p, &msg); err != nil {
				fpmApp.Logger.Errorf("parse renew message error: %v", err)
				return
			}
			if err := Renew(msg.Header.AppID, msg.Header.ProjID, msg.Payload.DeviceID, msg.Payload.Expire, p); err != nil {
				fpmApp.Logger.Errorf("do renew message error: %v", err)
				return
			}
		// https://shimo.im/docs/bJaoNiMc4yEfkRSt#anchor-bXxn
		case strings.HasSuffix(t, "message"):
		// https://shimo.im/docs/bJaoNiMc4yEfkRSt#anchor-dIeX
		case strings.HasSuffix(t, "feedback"):
		}
	})

	fpmApp.Run()

}

//Renew update the device active time
func Renew(appID string, projectID int64, deviceID string, expired int64, origin []byte) (err error) {
	device := fmt.Sprintf("%s:%s:%d:%s", redisPrefix, appID, projectID, deviceID)

	// if the device not exist, publish a online event
	isOk, err := check(device)
	if err != nil {
		return err
	}
	if !isOk {
		go func() {
			pushed, _ := publishOnlineEvent(device)
			log.Println("pushed: ", pushed)
		}()
	}
	isOk, err = renew(device, expired, origin)
	return
}

func publishOnlineEvent(deviceKey string) (bool, error) {
	appID, projectID, deviceID := splitDeviceKey(deviceKey)
	// publish the event
	fpmApp := fpm.Default()
	msg := message.RenewMessage{
		Header: &message.Header{
			Version:   10,
			NameSpace: "FPM.Lamp.Light",
			Name:      "Online",
			AppID:     appID,
			ProjID:    projectID,
			Source:    "MQTT",
		},
		Payload: &message.RenewPayload{
			DeviceID:  deviceID,
			Cgi:       deviceID,
			Timestamp: time.Now().Unix(),
		},
	}
	data, err := json.Marshal(&msg)
	if err != nil {
		return false, err
	}
	fpmApp.Execute("mqttclient.publish", &fpm.BizParam{
		"topic":   "$drm/" + appID + "/online",
		"payload": data,
	})
	return true, nil
}

func publishOfflineEvent(deviceKey string) (bool, error) {
	appID, projectID, deviceID := splitDeviceKey(deviceKey)
	// publish the event
	fpmApp := fpm.Default()
	msg := message.RenewMessage{
		Header: &message.Header{
			Version:   10,
			NameSpace: "FPM.Lamp.Light",
			Name:      "Offline",
			AppID:     appID,
			ProjID:    projectID,
			Source:    "MQTT",
		},
		Payload: &message.RenewPayload{
			DeviceID:  deviceID,
			Cgi:       deviceID,
			Timestamp: time.Now().Unix(),
		},
	}
	data, err := json.Marshal(&msg)
	if err != nil {
		return false, err
	}
	fpmApp.Execute("mqttclient.publish", &fpm.BizParam{
		"topic":   "$drm/" + appID + "/offline",
		"payload": data,
	})
	return true, nil
}

func genOnOfflineMessage() {

}

// split the redis key,
// like: drm:foo:bar
// it should return foo, bar
func splitDeviceKey(deviceKey string) (string, int64, string) {
	subStrs := strings.Split(deviceKey, ":")
	offset := 0
	if !strings.HasPrefix(deviceKey, redisPrefix) {
		offset = 1
	}
	appID, projectID, deviceID := subStrs[offset+1], subStrs[offset+2], subStrs[offset+3]
	id, _ := strconv.Atoi(projectID)
	return appID, int64(id), deviceID

}

// renew the device
func renew(device string, ex int64, origin []byte) (bool, error) {
	c, _ := fpm.Default().GetCacher()

	if err := c.SetString(device, string(origin), time.Duration(ex)*time.Second); err != nil {
		return false, err
	}

	return true, nil
}

// check the device
func check(device string) (bool, error) {
	c, _ := fpm.Default().GetCacher()
	return c.IsSet(device)
}
