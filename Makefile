all:
	make build-binary
	make build
	make run

build-binary:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o ox .

build:
	docker build -t gw/ox:latest .

run:
	docker run -d \
	--name ox \
	--log-driver=fluentd \
    --log-opt fluentd-address=139.59.19.30:24224 \
    --log-opt tag={{.Name}} \
	-p 80:80 \
	-p 9001:9001 \
	--rm \
	gw/ox:latest


delete-container:
	docker rm gw/ox:latest

delete-image:
	docker rmi gw/ox:latest

view-images:
	docker images

grpc:
	protoc -I ox_idl/ ox_idl/ox.proto --go_out=plugins=grpc:ox_idl

view-containers:
	docker ps -a