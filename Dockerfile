FROM golang:1.22.3-alpine as builder

WORKDIR /go/src/end

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN  go build -a -ldflags "-s -w -extldflags '-static'" -o evo-backend

FROM alpine:latest as prod
WORKDIR /app
COPY --from=builder /go/src/end/evo-backend /app/evo-backend
COPY ./config ./config
COPY ./contract ./contract
COPY ./data ./data
EXPOSE 2333

ENTRYPOINT ["/app/evo-backend"]