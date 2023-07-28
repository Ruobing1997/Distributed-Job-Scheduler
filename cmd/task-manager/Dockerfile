FROM golang:1.20 AS builder

WORKDIR /go/src/MoreFun_SuperNova

COPY go.mod go.sum ./

RUN go mod download

COPY . .
COPY cmd/task-manager/wait-for-it.sh /usr/wait-for-it.sh

RUN CGO_ENABLED=0 GOOS=linux go build -x -o manager cmd/task-manager/main.go

FROM alpine:latest

RUN apk add --no-cache bash

COPY --from=builder /usr/wait-for-it.sh /usr/wait-for-it.sh
COPY --from=builder /go/src/MoreFun_SuperNova/manager /app/manager

RUN chmod +x /usr/wait-for-it.sh

CMD ["/app/manager", "start"]