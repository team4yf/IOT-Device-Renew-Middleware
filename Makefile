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

run:
	docker run -p 5009:5009 -d yfsoftcom/fpm-iot-drm:${TAG}