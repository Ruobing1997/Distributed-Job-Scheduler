.PHONY: pull-redis up down build


pull-redis:
	docker pull redis:latest

up:
	docker-compose up -d

down:
	docker-compose down

build:
	docker-compose build