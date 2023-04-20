PROD_SERVER:=bubbles
HOST?=$(shell hostname)
PRJ_NAME=kae

ifeq ($(HOST), bubbles)
include .env.prod
export $(shell sed 's/=.*//' .env.prod)
endif

.PHONY: run open lint
lint:
	golangci-lint run

test:
	@ls *.go | entr -c -s 'go test -failfast -v ./*.go && notify "💚" || notify "🛑"'

single-run-test:
	go test -failfast -v *.go

coverage:
	@go test -cover ./...

.PHONY: check-vars
check-vars:
	echo u:$$ADMIN_USER p:$$ADMIN_PASS db:$$DB_FILE $$HOST

air:
	air

run:
	go run *.go

main-linux-amd64:
	GOARCH=amd64 GOOS=linux go build -ldflags "-w" -o $@ *.go

deploy: lint single-run-test clean main-linux-amd64 rsync
	ssh $(PROD_SERVER) "cd $(PRJ_NAME); make build restart prune"
	notify "🤘"

prune:
	docker system prune --force

sync-db-from-prod:
	cp ./sqlite.db /tmp
	scp $(PROD_SERVER):/data/$(PRJ_NAME)/sqlite.db ./sqlite.db
	@echo "Previous db here: /tmp/sqlite.db"

rsync:
	rsync --progress -avz ../$(PRJ_NAME) $(PROD_SERVER):. --exclude=.git

build:
	docker stop $(PRJ_NAME) 2>/dev/null ;\
	docker image rm --force $(PRJ_NAME) ;\
	docker build -t $(PRJ_NAME) .;\

restart:
	docker rm $(PRJ_NAME) ;\
	docker run \
		-v /data/$(PRJ_NAME):/data/$(PRJ_NAME) \
		--restart=unless-stopped \
		--name $(PRJ_NAME) \
		--network=$(DOCKER_NET) -p 9000:9000 \
		-d $(PRJ_NAME) \
		./main-linux-amd64 \
			--adminUser=$(ADMIN_USER) \
			--adminPass=$(ADMIN_PASS) \
			--dbFile=$(DB_FILE)

format:
	@ssh $(PROD_SERVER) 'docker ps --format "table {{.ID}}\t{{.Names}}\t{{.Networks}}\t{{.State}}\t{{.CreatedAt}}"'

clean: 
	rm -f $(PRJ_NAME) main-linux-amd64 main
