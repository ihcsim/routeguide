# server config
SERVER_PORT ?= 8080
FAULT_PERCENT ?= 0.3

# client config
SERVER_HOST ?= :$(SERVER_PORT)
GRPC_TIMEOUT ?= 20s
CLIENT_MODE ?= REPEATN
RPC_API ?= GetFeature
MAX_REPEAT ?= 15
ENABLE_LOAD_BALANCING ?= true
SERVER_IPV4 ?= 127.0.0.1:8080,127.0.0.1:8081,127.0.0.1:8082

server:
	go build -o ./cmd/server/server ./cmd/server/
	./cmd/server/server -port=$(SERVER_PORT) -fault-percent=$(FAULT_PERCENT)

client:
	go build -o ./cmd/client/client ./cmd/client/
	./cmd/client/client \
		-server=$(SERVER_HOST) \
		-timeout=$(GRPC_TIMEOUT) \
		-mode=$(CLIENT_MODE) \
		-api=$(RPC_API) \
		-n=$(MAX_REPEAT) \
		-enable-load-balancing=$(ENABLE_LOAD_BALANCING) \
		-server-ipv4=$(SERVER_IPV4)

l5d2:
	linkerd install --tls=optional | kubectl apply -f -

deploy:
	kubectl apply -f k8s-server.yaml
	sleep 15s
	kubectl apply -f k8s-client.yaml

mesh:
	linkerd inject --tls=optional k8s-server.yaml | kubectl apply -f -
	sleep 15s
	linkerd inject --tls=optional k8s-client.yaml | kubectl apply -f -

image:
	@eval $$(minikube docker-env) ; \
	docker build --rm -t routeguide .

proto:
	protoc -I proto/route_guide.proto --go_out=plugins=grpc:proto

clean:
	kubectl delete -f k8s-server.yaml
	kubectl delete -f k8s-client.yaml
