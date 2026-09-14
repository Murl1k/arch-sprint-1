FROM golang:1.27.1-alpine3.24 AS build
WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/temperature-api .

FROM alpine:3.24.1
COPY --from=build /out/temperature-api /usr/local/bin/temperature-api
EXPOSE 8081
ENTRYPOINT ["temperature-api"]
