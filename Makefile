build:
	@go build -v -o cascade.bin cmd/cascade/main.go
run:
	@DB_FILE=cascade.db JWT_SECRET=testing-token-do-not-use-in-production-insecure-token go run -v cmd/cascade/main.go
clean:
	@go clean -v