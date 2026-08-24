lint:
	golangci-lint run --fix $(TEST)

test: lint
	go test ./...