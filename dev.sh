# create binary
goreleaser --snapshot --rm-dist

# shutdown containers 
docker-compose down 

# run docker compose and rebuild images
docker-compose up --build