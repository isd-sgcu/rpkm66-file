proto:
	find . -name *.proto -exec protoc --proto_path=proto --go_out=. --go-grpc_out=. {} \;

test:
	go vet ./...
	go test  -v -coverpkg ./internal/... -coverprofile coverage.out -covermode count ./internal/...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

server:
	go run ./.

compose-up:
	docker-compose up -d

compose-down:
	docker-compose down

seed:
	go run ./. seed