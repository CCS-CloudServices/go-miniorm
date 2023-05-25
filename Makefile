run-test-env:
	./testing/docker/run-test-env.sh

test:
	go test ./...

lint:
	golangci-lint run