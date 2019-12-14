FROM golang as builder

WORKDIR /build/nombda
COPY . .

RUN go get -d -v ./...
RUN make build

FROM debian

COPY --from=builder nombda /usr/local/bin/nombda

EXPOSE 8080

CMD nombda
