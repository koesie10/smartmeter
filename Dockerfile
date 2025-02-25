FROM golang:1.23.6 as builder

RUN mkdir /app
WORKDIR /app
COPY . /app/
RUN CGO_ENABLED=0 go build -o bin/smartmeter ./cmd/smartmeter

FROM alpine:3.21.2
RUN apk --update add ca-certificates
RUN mkdir /app
WORKDIR /app
COPY --from=builder /app/bin/smartmeter /bin/smartmeter

ENTRYPOINT ["/bin/smartmeter"]
