FROM golang:1.13.6-alpine3.10 as builder

RUN mkdir /app
WORKDIR /app
COPY go.mod .
COPY go.sum .

RUN go mod download
COPY . .

WORKDIR /app/cmd
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o office-service

FROM scratch
COPY --from=builder /app/cmd/office-service .
CMD ["./office-service"]