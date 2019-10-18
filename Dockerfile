FROM golang:latest AS builder
ENV GO111MODULE=on

WORKDIR /app/

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -o main .

FROM scratch
COPY --from=builder /app ./
ENTRYPOINT ["./main"]
EXPOSE 8080
