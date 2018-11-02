GOVERSION := $(shell go version | cut -d" " -f3)
ELLD_ACCOUNT_PASS = ${ELLD_ACCOUNT_PASSWORD}

# Run some tests
test:
	cd blockchain && go test -v ./...
	cd blockcode && go test -v ./...
	cd accountmgr && go test -v ./...
	cd config && go test -v ./...
	cd crypto && go test -v ./...
	cd elldb && go test -v ./...
	cd miner && go test -v ./...
	cd node && go test -v ./...
	cd rpc && go test -v ./...
	cd util && go test -v ./...
	cd types && go test -v ./...

# Clean and format source code	
clean: 
	go vet ./... && gofmt -s -w .
	
# Ensure dep depencencies are in order
dep-ensure:
	dep ensure -v
	
# Create a release 
release:
	env GOVERSION=$(GOVERSION) goreleaser --snapshot --rm-dist
	
# Create a tagged release 
release-tagged:
	env GOVERSION=$(GOVERSION) goreleaser release --skip-publish --rm-dist

# Build an elld image 
build: 
	docker build -t elld-node -f ./docker/node/Dockerfile .
update: 
	docker build -t elld-node --no-cache -f ./docker/node/Dockerfile .
	
# Remove elld volume and container
destroy: 
	@echo "\033[0;31m[WARNING!]\033[0m You are about to remove 'elld' container and volumes. \n\
	Data (e.g. Accounts, Blockchain state, logs etc) in the volumes attached to an 'elld' \n\
	container will be lost forever."
	python ./scripts/confirm.py "docker rm -f -v elld"
	
# Starts elld client in a docker container
start:
	docker volume create elld-datadir
	docker run -d \
	 	--name elld \
		-e ELLD_ACCOUNT_PASSWORD=$(ELLD_ACCOUNT_PASS) \
		-p 0.0.0.0:9000:9000 \
		-p 0.0.0.0:8999:8999 \
		--mount "src=elld-datadir,dst=/root/.ellcrys" \
		elld-node
		
# Starts elld client in a docker container
# with the host data directory (~/.ellcrys) used as volume
start-hv:
	docker volume create elld-datadir
	docker run -d \
	 	--name elld \
		-e ELLD_ACCOUNT_PASSWORD=$(ELLD_ACCOUNT_PASS) \
		-p 0.0.0.0:9000:9000 \
		-p 0.0.0.0:8999:8999 \
		-v ~/.ellcrys:/root/.ellcrys \
		elld-node
		
# Gracefully stop the node
stop: 
	docker stop elld

# Restart a node	
restart:
	docker restart elld

remove: stop
	docker rm -f elld

# Follow logs
logs: 
	docker logs elld -f
	
# Attach to elld running locally
attach:
	elld attach
	
# Execute commands in the client's container
exec:
	docker exec -it elld bash -c "${c}"
	
# Starts a bash terminal
bash:
	docker exec -it elld bash

	
