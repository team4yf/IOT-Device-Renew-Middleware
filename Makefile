all: build push clean

TAG=
ifdef tag
	TAG=$(tag)
else
	TAG=latest
endif

build: 
	CGO_ENABLED=0 GOOS=linux go build -o main main.go
	docker build -t yfsoftcom/fpm-iot-drm:${TAG} .

push:
	docker push yfsoftcom/fpm-iot-drm:${TAG}

clean:
	rm -rf main
redis:
	docker run -p 6379:6379 -v $(PWD)/redis.conf:/usr/local/etc/redis/redis.conf -d redis:alpine3.11 redis-server /usr/local/etc/redis/redis.conf

run:
	docker run -p 5009:5009 -e "REDIS_HOST=192.168.159.102" -d yfsoftcom/fpm-iot-drm:${TAG}