FROM debian:stretch

# set working directory
WORKDIR /root/elld

# Install curl and other tools
RUN apt-get update
RUN apt-get install -y curl

# Download elld and decompress
RUN curl -L "https://storage.googleapis.com/elld-releases/elld_v0.0.0_linux_x86_64.tar.gz" > elld_v0.0.0_linux_x86_64.tar.gz
RUN tar -xvzf elld_v0.0.0_linux_x86_64.tar.gz

# Start client
EXPOSE 9000
EXPOSE 8999
ENTRYPOINT ./elld start --rpc -a "0.0.0.0:9000" --rpcaddress "0.0.0.0:8999" --pwd marvel