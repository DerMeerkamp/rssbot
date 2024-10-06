FROM golang:1.23-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o rssbot .

FROM alpine:edge

WORKDIR /app

COPY --from=build /app/rssbot .

RUN apk --no-cache add ca-certificates tzdata

ENTRYPOINT ["/app/rssbot"]