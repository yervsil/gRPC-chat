.PHONY: gen-proto
gen-proto:
	protoc ./chat.proto --go_out=./pkg/api/chat --go_opt=paths=source_relative --go-grpc_out=./pkg/api/chat --go-grpc_opt=paths=source_relative
