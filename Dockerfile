FROM golang:1.17-buster AS builder
WORKDIR /
COPY go.* ./
RUN go mod download
COPY main.go ./
RUN  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main

FROM gcr.io/distroless/base-debian10
WORKDIR /
COPY --from=builder /main /main
EXPOSE 8080
ENTRYPOINT ["/main"]