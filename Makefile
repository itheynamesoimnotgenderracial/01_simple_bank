
DB_URL=postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable

postgres:
	docker run --name postgres12 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -p 5432:5432 -d postgres:12-alpine

createdb:
	docker exec -it postgres12 createdb --username=root --owner=root simple_bank

dbnetworkkiller:
	docker exec -it postgres12 psql -U root -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'simple_bank';"

dropdb:
	docker exec -it postgres12 dropdb simple_bank

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migrateup1:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 1

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

migratedown1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down 1

migratereset:
	migrate -path db/migration -database "$(DB_URL)" force 0

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

test_rebuild:
	go test -v -cover -count=1 ./...

server:
	go run main.go

mock:
	mockgen  -package mockdb  -destination db/mock/store.go github.com/projects/go/01_simple_bank/db/sqlc Store

db_docs:
	dbdocs build docs/db.dbml

db_schema:
	dbml2sql --posgres -o docs/schema.sql docs/db.dbml

proto:
	rm -rf pb/*.go
	rm -rf docs/swagger/*.swagger.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt paths=source_relative \
	--openapiv2_out=docs/swagger --openapiv2_opt=allow_merge=true,merge_file_name=simple_bank  \
    proto/*.proto
	statik -src=./docs/swagger -dest=./docs
	
evans:
	docker run --rm -it \
	--network banknet \
	-v "$(CURDIR):/mount:ro" \
	ghcr.io/ktr0731/evans:latest \
	--path /mount/proto/ \
	--proto service_simple_bank.proto \
	--host api \
	--port 9090 \
	repl

# Tool → Module mapping
TOOLS = \
	github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway=github.com/grpc-ecosystem/grpc-gateway/v2 \
	github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2=github.com/grpc-ecosystem/grpc-gateway/v2 \
	google.golang.org/protobuf/cmd/protoc-gen-go=google.golang.org/protobuf \
	google.golang.org/protoc-gen-go-grpc=google.golang.org/protoc-gen-go-grpc


install-tools:
	@echo "Installing tools from go.mod..."
	@for toolmap in $(TOOLS); do \
		TOOL=$${toolmap%=*}; \
		MODULE=$${toolmap#*=}; \
		VERSION=$$(go list -m -f '{{.Version}}' $$MODULE); \
		echo "Installing $$TOOL@$${VERSION}..."; \
		go install $$TOOL@$$VERSION; \
	done




.PHONY: postgres createdb dropdb migrateup migratedown migratereset sqlc dbnetworkkiller server mock migrateup1 migratedown1 test_rebuild db_docs db_schema proto evans install-tools