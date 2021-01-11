
FROM golang:1.15-alpine

RUN apk update && apk add build-base

ENV GO111MODULE=on

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
ENV GOPROXY=https://proxy.golang.org

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
WORKDIR $GOPATH/src/togo
COPY . .

RUN GOOS=linux go build -o togo
CMD ["./togo"]
EXPOSE 5050