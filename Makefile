VERSION=v1.0.0
build:
	@cd cmd && go mod tidy && CGO_ENABLED=0 go build -ldflags="-s -w" -o ../bin/nginxgo.bin

build-linux:
	@cd cmd && go mod tidy && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ../bin/nginxgo.bin

docker-build:
	@docker build -t nginxgo:${VERSION} -f ./docker/Dockerfile .

docker-build-linux:
	@cd cmd && go mod tidy && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ../bin/nginxgo.bin
	@docker build -t nginxgo:${VERSION} -f ./docker/Dockerfile-linux .
