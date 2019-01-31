run_server:
	go run cmd/server/main.go

run_client:
	go run cmd/client/main.go

mesh:
	linkerd inject --tls=optional k8s.yaml | kubectl apply -f -

clean:
	kubectl delete -f server.yaml
	kubectl delete -f client.yaml

image_local:
	docker build --rm -t routeguide .

image_mkube:
	@eval $$(minikube docker-env) ; \
	docker build --rm -t routeguide .

proto:
	protoc -I proto/route_guide.proto --go_out=plugins=grpc:proto
