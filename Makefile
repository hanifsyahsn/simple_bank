create_db:
	docker exec -it simple_bank createdb -U postgres -O postgres simple_bank

drop_db:
	docker exec -it simple_bank dropdb -U postgres simple_bank

postgres:
	docker run --name simple_bank -e POSTGRES_PASSWORD=12345 -e POSTGRES_USER=postgres -e POSTGRES_DB=simple_bank -p 5432:5432 -v simple_bank:/var/lib/postgresql/data -d postgres:9.6

db_start:
	docker start simple_bank

db_stop:
	docker stop simple_bank

migrate_up:
	migrate -path db/migration -database "postgresql://postgres:12345@localhost:5432/simple_bank?sslmode=disable" -verbose up

migrate_down:
	migrate -path db/migration -database "postgresql://postgres:12345@localhost:5432/simple_bank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

.PHONY: create_db drop_db postgres db_start db_stop migrate_up migrate_down sqlc