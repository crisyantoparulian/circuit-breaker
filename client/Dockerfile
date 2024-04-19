FROM golang:1.20-alpine AS builder

WORKDIR /src

RUN mkdir app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . ./app

WORKDIR /src/app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags=jsoniter,netgo,nomsgpack -ldflags='-s -w -extldflags "-static"' -o client

# Serve the app
FROM public.ecr.aws/docker/library/alpine:latest

WORKDIR /app
COPY --from=builder /src/app/client /app/

EXPOSE 8080

ENTRYPOINT [ "/app/client" ]