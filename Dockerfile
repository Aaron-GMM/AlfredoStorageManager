# go_server/api/Dockerfile

FROM golang:1.24-alpine AS builder 

WORKDIR /app


COPY go.mod ./
COPY go.sum ./


RUN go mod download


COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o server .


FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/server .


COPY ./.env ./.env 


RUN chmod +x server


EXPOSE 8080


CMD ["./server"]
