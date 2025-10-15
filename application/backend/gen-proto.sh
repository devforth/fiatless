#!/bin/bash

INCLUDES="-I=./proto/tron -I/usr/lib -I./proto/googleapis"
FLAGS="--go_out=./pkg/proto/tron --go_opt=paths=source_relative --go-grpc_out=./pkg/proto/tron --go-grpc_opt=paths=source_relative"
CONTRACT_FLAGS="--go_out=./pkg/proto/tron/core --go_opt=paths=source_relative --go-grpc_out=./pkg/proto/tron/core --go-grpc_opt=paths=source_relative"

IMPORT_MAPPINGS="--go_opt=Mapi/api.proto=fiatless/pkg/proto/tron/api \
                 --go_opt=Mcore/Tron.proto=fiatless/pkg/proto/tron/core" \

protoc ${INCLUDES} ${FLAGS} ${IMPORT_MAPPINGS} ./proto/tron/core/*.proto ./proto/tron/api/*.proto
protoc ${INCLUDES} ${CONTRACT_FLAGS} ./proto/tron/core/contract/*.proto