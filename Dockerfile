FROM golang:1.9-alpine
RUN apk add --no-cache git
WORKDIR /go/src/github.com/nexeck/http-log/
RUN go get -u github.com/golang/dep/cmd/dep
COPY . .
RUN dep ensure
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/nexeck/http-log/server .
CMD ["./server"]