# Dockerfile
FROM golang:1.19-alpine AS buildStage

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go test -v --tags=unit ./...

RUN go build -o ./out/go-assessment .

# ===================

FROM alpine:3.16.2

COPY --from=buildStage /app/out/go-assessment /app/go-assessment

EXPOSE 2565

CMD ["/app/go-assessment"]