FROM golang:1.24.1

WORKDIR /app

COPY go.mod go.sum /
RUN go mod download

COPY . .

WORKDIR /app/services/validator/cmd
RUN go build -o main . && ls -la

EXPOSE 3000
CMD ["/app/services/validator/cmd/main"]