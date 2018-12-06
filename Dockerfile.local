FROM golang:1.10.5-stretch

ARG version=latest

# Move Elld binary from host
ADD ./dist/linux_amd64 /dist
RUN mv /dist/elld /usr/local/bin/elld

# Start client
EXPOSE 9000
EXPOSE 8999
ENTRYPOINT ["elld", "start", "-a", "0.0.0.0:9000"]