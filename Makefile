.PHONY: prepare build start stop delete e2e-prepare e2e-build e2e-start e2e-auth e2e-buy e2e-send load-test-build load-test-start

TARGET_DIR ?= .
ENV_FILE=$(TARGET_DIR)/.env
E2E_TEST=./tests/e2e

prepare:
	echo "#PostgreSQL DB" > $(ENV_FILE)
	echo "PG_SHARD_0=shard0" >> $(ENV_FILE)
	echo "PG_SHARD_1=shard1" >> $(ENV_FILE)
	echo "PG_SHARD_2=shard2" >> $(ENV_FILE)
	echo "PG_SHARD_3=shard3" >> $(ENV_FILE)
	echo "POSTGRES_USER=user" >> $(ENV_FILE)
	echo "POSTGRES_PASSWORD=secret" >> $(ENV_FILE)
	echo "DB=db" >> $(ENV_FILE)
	echo "" >> $(ENV_FILE)
	echo "#Backend" >> $(ENV_FILE)
	echo "ADDR=0.0.0.0:8080" >> $(ENV_FILE)
	echo "SECRET=super-secret-key" >> $(ENV_FILE)
	echo "ISSUER=merch" >> $(ENV_FILE)
	echo "TOKEN_TTL=30m" >> $(ENV_FILE)
	echo "CACHE_TTL=1h" >> $(ENV_FILE)

build:
	docker compose build

start:
	docker compose up -d

stop:
	docker compose down

delete:
	docker compose down -v

load-test-build:
	docker build -f ./pkg/k6/Dockerfile -t k6-test ./pkg/k6

load-test-start:
	docker run --network merch_network --rm k6-test

e2e-prepare:
	make prepare TARGET_DIR=tests

e2e-build:
	docker compose -f ./tests/docker-compose.yaml build

e2e-start:
	docker compose -f ./tests/docker-compose.yaml up -d

e2e-auth:
	go test -v $(E2E_TEST) -run ^TestE2EAuth

e2e-buy:
	go test -v $(E2E_TEST) -run ^TestE2EBuy

e2e-send:
	go test -v $(E2E_TEST) -run ^TestE2ETransfer

e2e-delete:
	docker compose -f ./tests/docker-compose.yaml down -v




