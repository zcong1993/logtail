FROM golang:1.14 AS build
WORKDIR /mnt
COPY . .
RUN CGO_ENABLED=0 go build -o ./bin/logtail ./cmd/main.go

FROM alpine:3.12
WORKDIR /opt
RUN apk add --no-cache ca-certificates
COPY --from=build /mnt/bin/* /usr/bin/
ENTRYPOINT ["logtail"]
