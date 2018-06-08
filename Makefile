proto:
	protoc --go_out=. pb/*.proto
test:
	go test ./...