FROM golang:1.21.0-alpine3.17 as builder
WORKDIR /build
COPY go.mod .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /main main.go

FROM scratch
COPY --from=builder main /bin/main
ENTRYPOINT [ "/bin/main" ]
