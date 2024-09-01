# use the official Golang image as the base image
FROM golang:alpine

WORKDIR /app

COPY go.sum go.mod ./

RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 8080
RUN chmod +x main

CMD ["./main"]