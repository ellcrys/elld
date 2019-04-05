GOVERSION := $(shell go version | cut -d" " -f3)

# Run tests
test:
	go test ./...
	
# Run tests (with ginkgo)
ginkgo:
	ginkgo ./...

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

# Create a release 
release-linux:
	cd $(GOPATH)/src/github.com/ellcrys/elld && \
	 git checkout ${b} && \
	 dep ensure -v && \
	 env GOVERSION=$(GOVERSION) goreleaser release --snapshot --rm-dist -f ".goreleaser.linux.yml"

# Build an elld image 
# use v=version to set Elld release tag
build: 
ifeq ('${v}','')
	docker build -t ellcrys/elld --no-cache .
else
	docker build --build-arg="version=tags/${v}" -t ellcrys/elld .
endif

# Rebuild the elld image
# use v=version to set Elld release tag
rebuild: 
ifeq ('${v}','')
	docker build -t ellcrys/elld --no-cache .
else
	docker build --build-arg="version=tags/${v}" -t ellcrys/elld --no-cache .
endif

build-local-linux: release-linux
	docker build -t ellcrys/elld --no-cache -f ./Dockerfile.local --no-cache .
	
# Starts elld client in a docker container
# with the host data directory (~/.ellcrys) used as volume
start:
	docker run -d \
	 	--name elld \
		-e ELLD_NET_VERSION="$(ELLD_NET_VERSION)" \
		-e ELLD_NODE_ADDRESS=$(ELLD_NODE_ADDRESS) \
		-e ELLD_NODE_BOOTSTRAPADDRS=$(ELLD_NODE_BOOTSTRAPADDRS) \
		-e ELLD_NODE_ACCOUNT=$(ELLD_NODE_ACCOUNT) \
		-e ELLD_NODE_PASSWORD="$(ELLD_NODE_PASSWORD)" \
		-e ELLD_RPC_ENABLED=$(ELLD_RPC_ENABLED) \
		-e ELLD_RPC_ADDRESS=$(ELLD_RPC_ADDRESS) \
		-e ELLD_RPC_USERNAME="$(ELLD_RPC_USERNAME)" \
		-e ELLD_RPC_PASSWORD="$(ELLD_RPC_PASSWORD)" \
		-e ELLD_DEBUG_CPUPROFILE=$(ELLD_DEBUG_CPUPROFILE) \
		-e ELLD_DEBUG_MEMPROFILE=$(ELLD_DEBUG_MEMPROFILE) \
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
	