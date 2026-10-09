export CGO_ENABLED := 0

VERSION ?= dev

.PHONY: web build test lint dev

web:
	cd web && npm ci && npm run build
	node --experimental-strip-types tools/sizes.ts

build: web
	go build -trimpath -ldflags '-s -w -X main.version=$(VERSION)' -o bin/wisp ./cmd/wisp

test: web
	go test ./...
	npm --prefix web test

lint:
	@test -z "$$(gofmt -l .)" || (gofmt -l .; exit 1)
	go vet ./...
	golangci-lint run ./...
	npm --prefix web run check

dev:
	@trap 'kill 0' INT TERM; \
	(cd web && npm run dev) & \
	go run ./cmd/wisp -dev & \
	wait
