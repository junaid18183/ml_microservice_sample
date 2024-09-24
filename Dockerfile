FROM golang:1.20.8-alpine as build

WORKDIR /src 
COPY * /src/

RUN go mod download
RUN go build -o /bin/ml

FROM registry.access.redhat.com/ubi8/ubi-micro

WORKDIR /usr/local/bin
COPY --from=build /bin/ml .
ENTRYPOINT ["/usr/local/bin/ml"]
