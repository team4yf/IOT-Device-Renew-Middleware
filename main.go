//Pacakge main the main entry

package main

// import fpm-server core & mqtt-client plugin & redis plugin
import (
	"encoding/json"
	"fmt"
	"github/team4yf/IOT-Device-Renew-Middleware/message"
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
		body := data.(map[string]interface{})
		t := body["topic"].(string)
		expiredData := body["payload"].(string)
		fpmApp.Logger.Debugf("redis topic: %s, payload: %+v", t, expiredData)

		// filter the other key event
		if !strings.HasPrefix(expiredData, redisBizPrefix+":"+redisPrefix) {
			return
		}

		keyType, _, _, _ := splitKey(expiredData)
		// the data includes device and msg, so we should add an other prefix
		// TODO:
		switch keyType {
		case "device":
			isOk, err := publishOfflineEvent(expiredData)
			if !isOk {
				fpmApp.Logger.Infof("publishOfflineEvent failed, error: %v", err)
			}
		case "msg":
			isOk, err := publishMsgTimeoutEvent(expiredData)
			if !isOk {
				fpmApp.Logger.Infof("publishMsgTimeoutEvent failed, error: %v", err)
			}
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
			msg := message.MsgMessage{}
			if err := json.Unmarshal(p, &msg); err != nil {
				fpmApp.Logger.Errorf("parse msg message error: %v", err)
				return
			}
			if err := Renew(msg.Header.AppID, msg.Header.ProjID, msg.Payload.MsgID, msg.Payload.Expire, p); err != nil {
				fpmApp.Logger.Errorf("do renew message error: %v", err)
				return
			}
		// https://shimo.im/docs/bJaoNiMc4yEfkRSt#anchor-dIeX
		case strings.HasSuffix(t, "feedback"):
			msg := message.D2SFeedbackMessage{}
			if err := json.Unmarshal(p, &msg); err != nil {
				fpmApp.Logger.Errorf("parse feedback message error: %v", err)
				return
			}
			key := fmt.Sprintf("%s:%s:%s:%d:%s", redisPrefix, "msg", msg.Header.AppID, msg.Header.ProjID, msg.Feedback.MsgID)
			if _, err := remove(key); err != nil {
				fpmApp.Logger.Errorf("do remove message error: %v", err)
				return
			}
		}
	})

	fpmApp.Run()

}

//Msg store a msg, notify a timeout event after timeout
func Msg(appID string, projectID int64, msgID string, expired int64, origin []byte) (err error) {
	key := fmt.Sprintf("%s:%s:%s:%d:%s", redisPrefix, "msg", appID, projectID, msgID)
	_, err = renew(key, expired, origin)
	return
}

//Renew update the device active time
func Renew(appID string, projectID int64, deviceID string, expired int64, origin []byte) (err error) {
	device := fmt.Sprintf("%s:%s:%s:%d:%s", redisPrefix, "device", appID, projectID, deviceID)

	// if the device not exist, publish a online event
	isOk, err := check(device)
	if err != nil {
		return err
	}
	if !isOk {
		go func() {
			publishOnlineEvent(device)
		}()
	}
	isOk, err = renew(device, expired, origin)
	return
}

func publishOnlineEvent(deviceKey string) (bool, error) {
	_, appID, projectID, deviceID := splitKey(deviceKey)
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
	_, appID, projectID, deviceID := splitKey(deviceKey)
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

func publishMsgTimeoutEvent(key string) (bool, error) {
	_, appID, projectID, msgID := splitKey(key)
	// publish the event
	fpmApp := fpm.Default()
	msg := message.MsgMessage{
		Header: &message.Header{
			Version:   10,
			NameSpace: "FPM.Lamp.Light",
			Name:      "Timeout",
			AppID:     appID,
			ProjID:    projectID,
			Source:    "MQTT",
		},
		Payload: &message.MsgPayload{
			MsgID:     msgID,
			Timestamp: time.Now().Unix(),
		},
	}
	data, err := json.Marshal(&msg)
	if err != nil {
		return false, err
	}
	fpmApp.Execute("mqttclient.publish", &fpm.BizParam{
		"topic":   "$drm/" + appID + "/timeout",
		"payload": data,
	})
	return true, nil
}

// split the redis key,
// like: bizPrefix:DRM:device:test:1:ff001
// it should return test, device, 1, ff001
func splitKey(deviceKey string) (string, string, int64, string) {
	subStrs := strings.Split(deviceKey, ":")
	offset := 0
	if !strings.HasPrefix(deviceKey, redisPrefix) {
		offset = 1
	}
	keyType, appID, projectID, deviceID := subStrs[offset+1], subStrs[offset+2], subStrs[offset+3], subStrs[offset+4]
	id, _ := strconv.Atoi(projectID)
	return keyType, appID, int64(id), deviceID

}

// renew the data
func renew(key string, ex int64, origin []byte) (bool, error) {
	c, _ := fpm.Default().GetCacher()

	if err := c.SetString(key, string(origin), time.Duration(ex)*time.Second); err != nil {
		return false, err
	}

	return true, nil
}

//remove remove a key
func remove(key string) (bool, error) {
	c, _ := fpm.Default().GetCacher()
	exists, err := c.IsSet(key)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}
	return c.Remove(key)
}

// check the device
func check(device string) (bool, error) {
	c, _ := fpm.Default().GetCacher()
	return c.IsSet(device)
}
