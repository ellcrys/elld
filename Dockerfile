FROM golang:1.10.5-stretch

WORKDIR /go/src/github.com/ellcrys

# clone elld repository
RUN git clone -b master https://github.com/ellcrys/elld.git

# get dependencies
WORKDIR elld
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
RUN make install-deps

# build from source
RUN go build
RUN mv elld /usr/local/bin/elld

# Start client
EXPOSE 9000
EXPOSE 8999
ENTRYPOINT ["elld", "start", "-a", "0.0.0.0:9000"]