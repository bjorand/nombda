FROM golang as builder

WORKDIR /build/nombda
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

FROM debian

COPY --from=builder /go/bin/nombda /usr/local/bin/nombda

EXPOSE 8080

CMD nombda
