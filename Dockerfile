FROM golang:1.24-bookworm as builder

WORKDIR /app

COPY ["./go.mod", "./go.sum", "./"]

RUN go mod download

COPY ./ ./

RUN go build -o dutil .

FROM gcr.io/distroless/base-debian12

COPY --from=builder /app/dutil /
ENTRYPOINT ["/dutil"]