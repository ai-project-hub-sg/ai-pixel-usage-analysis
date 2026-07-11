.PHONY: test web-build build build-linux verify

test:
	cd web && npm run typecheck
	cd web && npm test -- --run
	go test ./...
web-build:
	cd web && npm run build
build:
	$(MAKE) web-build
	go build -o dist/ai-pixel-usage-analysis ./cmd/ai-pixel-usage-analysis
build-linux:
	$(MAKE) web-build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/ai-pixel-usage-analysis-linux-amd64 ./cmd/ai-pixel-usage-analysis
verify:
	$(MAKE) test
	go vet ./...
	$(MAKE) build-linux
