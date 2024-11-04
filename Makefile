dsn=user:password@tcp(db:3306)/data?charset=utf8&parseTime=True&loc=Local
migrationPath=./migrations
driver=mysql

run: build

build:
	@docker compose exec app go build -o ./cmd/server/main.go .

seed:
	@docker compose run app go run cmd/seed/main.go

db-status:
	@docker compose run app sh -c 'GOOSE_DRIVER="$(driver)" GOOSE_DBSTRING="$(dsn)" goose -dir="$(migrationPath)" status'

up:
	@docker compose run app sh -c 'GOOSE_DRIVER="$(driver)" GOOSE_DBSTRING="$(dsn)" goose -dir="$(migrationPath)" up'

down:
	@docker compose run app sh -c 'GOOSE_DRIVER="$(driver)" GOOSE_DBSTRING="$(dsn)" goose -dir="$(migrationPath)" down'

reset:
	@docker compose run app sh -c 'GOOSE_DRIVER="$(driver)" GOOSE_DBSTRING="$(dsn)" goose -dir="$(migrationPath)" reset'