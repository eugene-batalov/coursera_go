FROM golang:1.8

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...

CMD ["bash"]