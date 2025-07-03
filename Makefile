# Run all tests
.PHONY: test
test:
	go test -v ./...

# Run the application
.PHONY: run
run:
	go run main.go