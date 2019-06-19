FROM golang:alpine

RUN apk update && apk upgrade && \
    apk add --no-cache ca-certificates && \
    apk add git

ENV GOBIN /go/bin

# build directories
RUN mkdir /app
RUN mkdir /go/src/app
ADD . /go/src/app
WORKDIR /go/src/app

# Go dep!
RUN go get -u github.com/golang/dep/...
RUN dep ensure

# Build my app
RUN go build -o /app/main .
CMD ["/app/main"]