FROM golang:latest AS builder
COPY main.go /app/
WORKDIR /app/
RUN go get github.com/gorilla/mux
RUN go get github.com/gorilla/handlers
RUN CGO_ENABLED=0 GOOS=linux go build -a -o main .

FROM scratch
COPY --from=builder /app ./
ENTRYPOINT ["./main"]
EXPOSE 8080
