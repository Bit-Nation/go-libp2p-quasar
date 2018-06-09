install:
	gx install
proto:
	protoc --go_out=. pb/*.proto
test:
	go test ./...
deps_hack:
	gx-go rw
deps_hack_revert:
	gx-go uw