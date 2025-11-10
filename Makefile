create_db:
	docker exec -it simple_bank createdb -U postgres -O postgres simple_bank

drop_db:
	docker exec -it simple_bank dropdb -U postgres simple_bank

postgres:
	docker run --name simple_bank --network bank-network -e POSTGRES_PASSWORD=12345 -e POSTGRES_USER=postgres -e POSTGRES_DB=simple_bank -p 5432:5432 -v simple_bank:/var/lib/postgresql/data -d postgres:9.6

db_start:
	docker start simple_bank

db_stop:
	docker stop simple_bank

migrate_up:
	migrate -path db/migration -database "postgresql://postgres:12345@localhost:5432/simple_bank?sslmode=disable" -verbose up

migrate_down:
	migrate -path db/migration -database "postgresql://postgres:12345@localhost:5432/simple_bank?sslmode=disable" -verbose down

migrate_up1:
	migrate -path db/migration -database "postgresql://postgres:12345@localhost:5432/simple_bank?sslmode=disable" -verbose up 1

migrate_down1:
	migrate -path db/migration -database "postgresql://postgres:12345@localhost:5432/simple_bank?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

test_coverage:
	go test -v -coverprofile=coverage.out ./...

coverage_report:
	go tool cover -func=coverage.out

coverage_report_view:
	go tool cover -html=coverage.out

server:
	go run main.go

mock:
	mockgen -package mockdb --destination db/mock/store.go github.com/hanifsyahsn/simple_bank/db/sqlc Store

DIRE ?=
NAME ?=

migrate_create:
	migrate create -ext sql -dir $(DIRE) -seq $(NAME)

PACKAGE ?=

test_package:
	go test -v -count=1 $(PACKAGE)

.PHONY: create_db drop_db postgres db_start db_stop migrate_up migrate_down sqlc, test, test_coverage, coverage_report, coverage_report_view, server, mock, migrate_create, migrate_up1, migrate_down1, test_package