FROM golang:1.24.3-alpine

RUN apk add --no-cache \
    gcc \
    g++ \
    musl-dev \
    git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main ./cmd/main.go

RUN mkdir -p static/uploads

EXPOSE 8080

CMD ["./main"]