all:
	CGO_ENABLED=0 GOOS=linux go build -ldflags '-extldflags "-static"' .
	docker build -t quay.io/jsafrane/caas .
