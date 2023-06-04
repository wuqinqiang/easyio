
.PHONY: server
build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/main server.go



.PHONY: server
server:
	docker-compose build && docker-compose up