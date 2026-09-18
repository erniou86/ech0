FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY main.go ./
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o ech0 .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /app/ech0 .
ENV PORT=8080 DB_PATH=/data/data.db
VOLUME ["/data"]
EXPOSE 8080
CMD ["./ech0"]
