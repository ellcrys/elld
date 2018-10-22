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

# Build an elld image 
dockerize: 
	docker build -t elld-node -f ./docker/node/Dockerfile .
dockerize-force: 
	docker build -t --no-cache elld-node -f ./docker/node/Dockerfile .
	
# Start a node
start:
	docker volume create elld-datadir
	docker run -d \
	 	--name elld \
		-e ELLD_ACCOUNT_PASSWORD=$(ELLD_ACCOUNT_PASS) \
		-p 0.0.0.0:9000:9000 \
		-p 0.0.0.0:8999:8999 \
		--mount "src=elld-datadir,dst=/root/.ellcrys" \
		elld-node
		
# Gracefully stop the node
stop: 
	docker stop elld

remove: stop
	docker rm -f elld

# Attach to elld running locally
attach:
	elld attach
	
# Execute commands in the client's container
exec:
	docker exec -it elld bash
	
# Start elld in console mode with RPC enabled
run-console: 
	./dist/darwin_amd64/elld console --rpc
	
