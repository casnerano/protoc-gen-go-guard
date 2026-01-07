LOCAL_BIN := ${CURDIR}/bin
EXAMPLE_DIR := ${CURDIR}/example
GO_COVER_EXCLUDE := "example|.*\.pb\.go"

.PHONY: download-bin-deps
download-bin-deps:
	ls $(LOCAL_BIN)/golangci-lint &> /dev/null || GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.7.2
	ls $(LOCAL_BIN)/buf &> /dev/null || GOBIN=$(LOCAL_BIN) go install github.com/bufbuild/buf/cmd/buf@latest
	ls $(LOCAL_BIN)/protoc-gen-go &> /dev/null || GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
	ls $(LOCAL_BIN)/protoc-gen-go-grpc &> /dev/null || GOBIN=$(LOCAL_BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0

.PHONY: generate-guard-proto
generate-guard-proto:
	$(LOCAL_BIN)/buf generate --config ${CURDIR}/buf.yaml --template ${CURDIR}/buf.gen.yaml

.PHONY: generate-example-proto
generate-example-proto:
	$(LOCAL_BIN)/buf generate --config ${EXAMPLE_DIR}/buf.yaml --template ${EXAMPLE_DIR}/buf.gen.yaml

.PHONY: build-protoc-gen-go-guard
build-protoc-gen-go-guard:
	go build -o ${LOCAL_BIN}/protoc-gen-go-guard ./cmd/protoc-gen-go-guard

.PHONY: generate
generate: download-bin-deps generate-guard-proto build-protoc-gen-go-guard generate-example-proto
	go mod tidy

.PHONY: clean
clean:
	rm -rf $(LOCAL_BIN)
	rm -rf ./proto/*.pb.go
	rm -rf ./example/pb/*.pb.go

.PHONY: test
test:
	go test -count=1 -tags=integration ./...

.PHONY: test-cover
test-cover:
	go test -count=1 -cover -coverprofile=cover.temp.out -covermode=atomic ./...
	grep -vE ${GO_COVER_EXCLUDE} cover.temp.out > cover.out && rm cover.temp.out
	go tool cover -func cover.out

.PHONY: test-cover-with-html
test-cover-with-html: test-cover
	go tool cover -html=cover.out

.PHONY: lint
lint:
	$(LOCAL_BIN)/golangci-lint run ./...