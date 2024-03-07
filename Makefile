proto:
	protoc -I./proto --go_out=./ --go-grpc_out=./ **/*.proto

run:
	docker compose build node1; \
	docker compose up --build \

.PHONY: proto run