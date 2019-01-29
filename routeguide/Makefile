run_server:
	go run cmd/server/main.go

run_client:
	go run cmd/client/main.go

image:
	docker build --rm -t routeguide .

proto:
	protoc -I proto/route_guide.proto --go_out=plugins=grpc:proto
