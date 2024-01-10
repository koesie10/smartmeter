FROM golang:1.21.6 as builder

RUN mkdir /app
WORKDIR /app
COPY . /app/
RUN CGO_ENABLED=0 go build -o smartmeter ./cmd/smartmeter

FROM alpine:3.19.1
RUN apk --update add ca-certificates
RUN mkdir /app
WORKDIR /app
COPY --from=builder /app/smartmeter /bin/smartmeter

ENTRYPOINT ["/bin/smartmeter"]
