all: build push clean

PROJECTNAME=$(shell basename "$(PWD)")
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin

TAG=
ifdef tag
	TAG=$(tag)
else
	TAG=latest
endif

PASS=
ifdef pass
	PASS=$(pass)
endif

dev:
	go build -o $(GOBIN)/app ./main.go && $(GOBIN)/app

build: 
	# CGO_ENABLED=0 GOOS=linux go build -o main main.go
	docker build -t yfsoftcom/fpm-iot-drm:${TAG} .

push:
	docker push yfsoftcom/fpm-iot-drm:${TAG}

clean:
	rm -rf main

test-renew:
	# 如果要复制到 shell 中运行，需要此处的 $$ 换成 $ ， makfile中自动转义了 $, 所以这里放了2个 $
	mosquitto_pub -h open.yunplus.io -t '$$drm/test/renew' -m '{"header":{"v":10,"ns":"FPM.Lamp.Light","name":"Renew","appId":"test","projId":2,"source":"MQTT"},"payload":{"sn":"ff0111","expire":10,"cgi":"ff0111","timestamp":1594829482984}}' -u "fpmuser" -P '${PASS}'

sub-offline:
	mosquitto_sub -h open.yunplus.io -t '$$drm/test/offline'  -u "fpmuser" -P '${PASS}'

redis:
	docker run -p 6379:6379 -v $(PWD)/redis.conf:/usr/local/etc/redis/redis.conf -d redis:alpine3.11 redis-server /usr/local/etc/redis/redis.conf

run:
	docker run -p 3009:3009 -e "REDIS_HOST=172.19.69.130" -d yfsoftcom/fpm-iot-drm:${TAG}