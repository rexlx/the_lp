FROM golang:1.24-alpine as builder

WORKDIR /app
COPY go.mod go.sum ./
COPY . .
RUN go mod download
RUN go build -mod=readonly -o thelp .
RUN chmod +x thelp

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/thelp .
COPY scripts/ ./scripts/

EXPOSE 8080

# cmd should run ./thelp -db "user=rxlx password=thereISnosp0)n host=192.168.86.120 dbname=tags"
CMD ["./thelp", "-db", "user=rxlx password=thereISnosp0)n host=192.168.86.120 dbname=tags"]


