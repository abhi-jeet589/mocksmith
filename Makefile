.PHONY: up down logs build rebuild clean ps

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f api

build:
	docker compose build

rebuild:
	docker compose up -d --build

ps:
	docker compose ps

clean:
	docker compose down -v
