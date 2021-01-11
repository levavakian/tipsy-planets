FROM golang:1.14

WORKDIR /root

RUN apt-get update && apt-get install -y git sudo curl wget nano unzip python3-pip gnupg
RUN wget -qO - https://www.mongodb.org/static/pgp/server-4.2.asc | sudo apt-key add -
RUN curl -sL https://deb.nodesource.com/setup_10.x | bash - && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get -y install nodejs

RUN adduser apps
RUN echo "apps     ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers
RUN mkdir -p /home/apps/mmr-web && chown apps:apps /home/apps
COPY client /home/apps/tipsy-planets/client
COPY server /go/src/tipsy-planets/server
ENV GO111MODULE=on
WORKDIR /home/apps/tipsy-planets/client
RUN npm install
RUN npm run build
WORKDIR /go/src/tipsy-planets/server
RUN go get && go install

USER apps
WORKDIR /home/apps/tipsy-planets
