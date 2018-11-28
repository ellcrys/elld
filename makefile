GOVERSION := $(shell go version | cut -d" " -f3)
LADDRESS = ${ELLD_LADDRESS}
ADDNODE = ${ELLD_ADDNODE}
ACCOUNT_PASSWORD = ${ELLD_ACCOUNT_PASSWORD}
ACCOUNT = ${ELLD_ACCOUNT}
RPC_ON = ${ELLD_RPC_ON}
RPC_ADDRESS = ${ELLD_RPC_ADDRESS}
RPC_USERNAME = ${ELLD_RPC_USERNAME}
RPC_PASSWORD = ${ELLD_RPC_PASSWORD}
CPU_PROFILING_ON = ${ELLD_CPU_PROFILING_ON}
MEM_PROFILING_ON = ${ELLD_MEM_PROFILING_ON}

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

# Install source code and binary dependencies
deps: dep-ensure
	go get github.com/gobuffalo/packr/packr

# Create a release 
release:
	env GOVERSION=$(GOVERSION) goreleaser --snapshot --rm-dist
	
# Create a tagged release 
release-tagged:
	env GOVERSION=$(GOVERSION) goreleaser release --skip-publish --rm-dist

# Build an elld image 
build: 
	docker build -t ellcrys/elld .

# Rebuild the elld image
rebuild: 
	docker build -t ellcrys/elld --no-cache .
	
# Starts elld client in a docker container
# with the host data directory (~/.ellcrys) used as volume
start:
	docker run -d \
	 	--name elld \
		-e ELLD_LADDRESS=$(LADDRESS) \
		-e ELLD_ADDNODE=$(ADDNODE) \
		-e ELLD_ACCOUNT_PASSWORD=$(ACCOUNT_PASSWORD) \
		-e ELLD_ACCOUNT=$(ACCOUNT) \
		-e ELLD_RPC_ON=$(RPC_ON) \
		-e ELLD_RPC_ADDRESS=$(RPC_ADDRESS) \
		-e ELLD_RPC_USERNAME=$(RPC_USERNAME) \
		-e ELLD_RPC_PASSWORD=$(RPC_PASSWORD) \
		-e ELLD_CPU_PROFILING_ON=$(CPU_PROFILING_ON) \
		-e ELLD_MEM_PROFILING_ON=$(MEM_PROFILING_ON) \
		-p 0.0.0.0:9000:9000 \
		-p 0.0.0.0:8999:8999 \
		-v ~/.ellcrys:/root/.ellcrys \
		ellcrys/elld
		
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
	docker exec -it elld bash -c "elld attach"
	
# Execute commands in the client's container
exec:
	docker exec -it elld bash -c "${c}"
	
# Starts a bash terminal
bash:
	docker exec -it elld bash
	
# Remove elld volume and container
destroy: 
	@echo "\033[0;31m[WARNING!]\033[0m You are about to remove 'elld' container and volumes. \n\
	Data (e.g. Accounts, Blockchain state, logs etc) in the volumes attached to an 'elld' \n\
	container will be lost forever."
	python ./scripts/confirm.py "docker rm -f -v elld && docker volume remove -f elld-datadir"
	