FROM golang:1.24-alpine

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./

RUN go mod download

ENV CGO_ENABLED=1

COPY . .

RUN go build -o main .

EXPOSE 8081

CMD ["./main"]
