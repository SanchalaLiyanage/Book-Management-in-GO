FROM golang:1.21 as builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o book-api .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/book-api .
COPY data/books.json ./data/books.json

EXPOSE 8080
CMD ["./book-api"]