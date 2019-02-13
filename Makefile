server:
	go run cmd/server/main.go

client:
	go run cmd/client/main.go

l5d2:
	linkerd install --tls=optional | kubectl apply -f -

mesh:
	linkerd inject --tls=optional k8s.yaml | kubectl apply -f -

image:
	@eval $$(minikube docker-env) ; \
	docker build --rm -t routeguide .

proto:
	protoc -I proto/route_guide.proto --go_out=plugins=grpc:proto

clean:
	kubectl delete -f server.yaml
	kubectl delete -f client.yaml
	linkerd install --tls=optional | kubectl delete -f -
