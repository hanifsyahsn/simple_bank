to run test:
- go test -v -run ^test_function_name(such TestCreateAccount)$ test_location(such ./db/sqlc)
to run package test with coverage:
- go test -v -cover test_location(such ./db/sqlc)
to run package test with coverage and report
- go test -v -coverprofile=coverage.out test_location(such ./db/sqlc)
- go tool cover -func=coverage.out
