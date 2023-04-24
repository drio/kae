PROD_SERVER:=bubbles
HOST?=$(shell hostname)
PRJ_NAME=kae
DOCKER_NET=pihole_default

# Load env variables when in prod
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

coverage/html:
	go test -v -cover -coverprofile=c.out
	go tool cover -html=c.out

.PHONY: check-vars
check-vars:
	echo u:$$KAE_USER p:$$KAE_PASS db:$$KAE_DB ds:$$KAE_DELAY_SECS p:$$PORT

air:
	air

run:
	go run .

main-linux-amd64:
	GOARCH=amd64 GOOS=linux go build -ldflags "-w" -o $@ .

clean: 
	rm -f $(PRJ_NAME) main-linux-amd64 main

deploy: lint single-run-test clean main-linux-amd64 rsync
	ssh $(PROD_SERVER) "cd $(PRJ_NAME); make docker/build docker/restart docker/prune"
	notify "🤘"

rsync:
	rsync --progress -avz ../$(PRJ_NAME) $(PROD_SERVER):. --exclude=.git

build: docker/stop docker/image docker/build

docker/restart: docker/stop docker/rm docker/run

docker/prune:
	docker system prune --force

docker/stop:
	docker stop $(PRJ_NAME) 2>/dev/null

docker/image:
	docker image rm --force $(PRJ_NAME) 

docker/build:
	docker build -t $(PRJ_NAME) .

docker/rm:
	docker rm $(PRJ_NAME)

docker/run:
	docker run \
		--detach \
		--volume=/data/$(PRJ_NAME):/data/$(PRJ_NAME) \
		--env-file=./.env.prod \
		--restart=on-failure:5 \
		--name $(PRJ_NAME) \
		--network=$(DOCKER_NET) \
		$(PRJ_NAME) \
		./main-linux-amd64 -delaySecs=$(KAE_DELAY_SECS) 

docker/format:
	docker ps --format "table {{.ID}}\t{{.Names}}\t{{.Networks}}\t{{.State}}\t{{.CreatedAt}}"
