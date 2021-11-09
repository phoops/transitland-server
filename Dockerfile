FROM golang:1.16 as builder
WORKDIR $GOPATH/src/github.com/interline-io/transitland-server
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go test ./...
RUN CGO_ENABLED=0 go build -o /tlserver github.com/interline-io/transitland-server/cmd/tlserver

FROM alpine
RUN apk add --no-cache tzdata
WORKDIR /app
COPY --from=builder /tlserver /app
ENTRYPOINT [ "./tlserver" ]
CMD ["server", "-playground", "-auth=admin"]
