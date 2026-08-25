lint:
	golangci-lint run --fix $(TEST)

test:
	go test ./...