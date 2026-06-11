.PHONY: test test-race test-fuzz test-cover lint test-integration evaluate evaluate-ingest-smoke scrub-emls

test:
	go test ./... -count=1

evaluate:
	go run ./cmd/evaluate

evaluate-gate:
	go run ./cmd/evaluate -generated 10000 -seed 42 -json docs/milestones/phase-00-matching-engine/gate-results.json

evaluate-ingest-smoke:
	go run ./cmd/ingest-evaluate -json test/testdata/ingest/fixture-smoke-results.json

scrub-emls:
	go run ./cmd/scrub-eml -in emls -out test/testdata/email/merchants

test-race:
	go test ./... -race -count=1

FUZZTIME_SHORT ?= 5s

test-fuzz:
	go test ./internal/match/ -fuzz=FuzzEngine_NoFalsePositiveOnCurrencyMismatch -fuzztime=$(FUZZTIME_SHORT) -count=1
	go test ./internal/match/ -fuzz=FuzzMerchantResolver_Bounded -fuzztime=$(FUZZTIME_SHORT) -count=1
	go test ./internal/match/similarity/ -fuzz=FuzzJaroWinkler_Bounded -fuzztime=$(FUZZTIME_SHORT) -count=1
	go test ./internal/match/similarity/ -fuzz=FuzzTokenSetRatio_Bounded -fuzztime=$(FUZZTIME_SHORT) -count=1
	go test ./internal/synth/ -fuzz=FuzzGenerator_NoiseBounds -fuzztime=$(FUZZTIME_SHORT) -count=1
	go test ./internal/mail/ -fuzz=FuzzParseMaildirMessage_NoPanic -fuzztime=$(FUZZTIME_SHORT) -count=1
	go test ./internal/mail/ -fuzz=FuzzParseRFC822_NoPanic -fuzztime=$(FUZZTIME_SHORT) -count=1
	go test ./internal/extract/ -fuzz=FuzzParsePoundsToMinor_NoPanic -fuzztime=$(FUZZTIME_SHORT) -count=1
	go test ./internal/extract/ -fuzz=FuzzParseJSONLDTotal_NoPanic -fuzztime=$(FUZZTIME_SHORT) -count=1
	go test ./internal/extract/ -fuzz=FuzzParseHTMLTableTotal_NoPanic -fuzztime=$(FUZZTIME_SHORT) -count=1
	go test ./internal/extract/ -fuzz=FuzzExtractVisibleText_NoPanic -fuzztime=$(FUZZTIME_SHORT) -count=1
	go test ./internal/extract/ -fuzz=FuzzExtractPDFText_NoPanic -fuzztime=$(FUZZTIME_SHORT) -count=1

test-cover:
	go test ./... -coverprofile=coverage.out -count=1
	go tool cover -func=coverage.out | tail -1

test-integration:
	go test -tags=integration ./test/... -count=1

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...
