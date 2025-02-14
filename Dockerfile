FROM golang:1.23.5-alpine3.21 AS builder

ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /go/src/backend

RUN apk --update --no-cache add ca-certificates gcc libtool make musl-dev protoc git

COPY . /go/src/backend

RUN go mod download

RUN go build -o backend ./cmd/backend/main.go

FROM scratch

COPY --from=builder /go/src/backend/backend backend
COPY --from=builder /go/src/backend/internal/migrations/ /migrations
COPY --from=builder /go/src/backend/.env ./

EXPOSE 8080

ENTRYPOINT ["./backend"]