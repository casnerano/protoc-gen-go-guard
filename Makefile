LOCAL_BIN := $(CURDIR)/bin
EXAMPLE_DIR := $(CURDIR)/example

.PHONY: download-bin-deps
download-bin-deps:
	ls $(LOCAL_BIN)/buf &> /dev/null || GOBIN=$(LOCAL_BIN) go install github.com/bufbuild/buf/cmd/buf@latest
	ls $(LOCAL_BIN)/protoc-gen-go &> /dev/null || GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
	ls $(LOCAL_BIN)/protoc-gen-go-grpc &> /dev/null || GOBIN=$(LOCAL_BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0

.PHONY: generate-rbac-proto
generate-rbac-proto:
	$(LOCAL_BIN)/buf generate --config ${CURDIR}/buf.yaml --template ${CURDIR}/buf.gen.yaml

.PHONY: generate-example-proto
generate-example-proto:
	$(LOCAL_BIN)/buf generate --config ${EXAMPLE_DIR}/buf.yaml --template ${EXAMPLE_DIR}/buf.gen.yaml

.PHONY: build-protoc-gen-go-rbac
build-protoc-gen-go-rbac:
	go build -o ${LOCAL_BIN}/protoc-gen-go-rbac ./cmd/protoc-gen-go-rbac

.PHONY: generate
generate: download-bin-deps generate-rbac-proto build-protoc-gen-go-rbac generate-example-proto
	go mod tidy

.PHONY: clean
clean:
	rm -rf $(LOCAL_BIN)
	rm -rf $(CURDIR)/proto/*.pb.go
	rm -rf $(CURDIR)/example/pb/*.pb.go

example-run:
	go run ${EXAMPLE_DIR}/main.go

example-test:
	go test ${EXAMPLE_DIR}/...
