# create binary
env GOVERSION=v0.1.0 goreleaser --snapshot --rm-dist

# shutdown containers 
docker-compose down 

# run docker compose and rebuild images
docker-compose up --build