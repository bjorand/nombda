FROM golang as builder

WORKDIR /build/nombda
COPY . .

RUN go get -d -v ./...
RUN make build

FROM debian

RUN apt-get update && \
    apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    git \
    gnupg2 \
  	lsb-release && \
    curl -fsSL https://download.docker.com/linux/debian/gpg | apt-key add - && \
    echo \
    "deb https://download.docker.com/linux/debian \
    $(lsb_release -cs) \
    stable" > /etc/apt/sources.list.d/docker.list && \
    apt-get update && \
    apt-get install -y docker-ce-cli

COPY --from=builder /build/nombda/nombda /usr/local/bin/nombda

EXPOSE 8080

CMD nombda
