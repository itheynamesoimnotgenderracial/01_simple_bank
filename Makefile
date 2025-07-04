postgres:
	docker run --name postgres12 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -p 5432:5432 -d postgres:12-alpine

createdb:
	docker exec -it postgres12 createdb --username=root --owner=root simple_bank

dbnetworkkiller:
	docker exec -it postgres12 psql -U root -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'simple_bank';"

dropdb:
	docker exec -it postgres12 dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:1jbv9ENV4xzREdzb3tRI@simplebank.ct08gqmqu1ia.ap-southeast-1.rds.amazonaws.com/simple_bank" -verbose up

migrateup1:
	migrate -path db/migration -database "postgresql://root:1jbv9ENV4xzREdzb3tRI@simplebank.ct08gqmqu1ia.ap-southeast-1.rds.amazonaws.com/simple_bank" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgresql://root:1jbv9ENV4xzREdzb3tRI@simplebank.ct08gqmqu1ia.ap-southeast-1.rds.amazonaws.com/simple_bank" -verbose down

migratedown1:
	migrate -path db/migration -database "postgresql://root:1jbv9ENV4xzREdzb3tRI@simplebank.ct08gqmqu1ia.ap-southeast-1.rds.amazonaws.com/simple_bank" -verbose down 1

migratereset:
	migrate -path db/migration -database "postgresql://root:1jbv9ENV4xzREdzb3tRI@simplebank.ct08gqmqu1ia.ap-southeast-1.rds.amazonaws.com/simple_bank" force 0

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

.PHONY: postgres createdb dropdb migrateup migratedown migratereset sqlc dbnetworkkiller server mock migrateup1 migratedown1 test_rebuild