FROM "golang:stretch"

# set working directory
WORKDIR /app

# add druid binary to working directory
ADD ./dist/linux_amd64 /app

# expose druid port
EXPOSE 9000

# set bootstrap node addresses
ARG addnode
ENV addnode ${addnode}
ARG seed
ENV seed ${seed}

CMD ./druid start --dev -a 0.0.0.0:9000 -s ${seed} ${addnode}
