.PHONY: test test-race test-fuzz test-cover lint test-integration evaluate

test:
	go test ./... -count=1

evaluate:
	go run ./cmd/evaluate

test-race:
	go test ./... -race -count=1

test-fuzz:
	go test ./internal/match/ -fuzz=Fuzz -fuzztime=10s -count=1

test-cover:
	go test ./... -coverprofile=coverage.out -count=1
	go tool cover -func=coverage.out | tail -1

test-integration:
	go test -tags=integration ./test/... -count=1

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...
