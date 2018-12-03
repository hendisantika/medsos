FROM golang:latest as builder
RUN mkdir /src
ADD . /src
WORKDIR /src
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o medsos .

FROM alpine:latest
RUN mkdir /app
COPY --from=builder /src/medsos /app
WORKDIR /app
CMD ["./medsos"]
