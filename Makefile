run-stack:
	@docker compose up -d

stop-stack:
	@docker compose down

build:
	@go build -o toko-mas-jawa-backend ./cmd/

run:
	@./toko-mas-jawa-backend