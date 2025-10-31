-include local.mk

build-frontend:
	docker image rm videocall-frontend -f && docker compose up frontend

build-backend:
	docker image rm videocall-backend -f && docker compose build backend

run:
	docker compose --profile dev up --remove-orphans