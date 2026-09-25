FROM golang:1.26 AS builder
WORKDIR /src

# Separate layer for deps so a source-only change doesn't re-download modules.
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY static ./static
COPY templates ./templates
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /score .

# Static binary on distroless: CA certs for Google APIs, no shell, runs as nonroot.
# Templates and the favicon are embedded in the binary.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /score /score

# Cloud Run sets PORT (8080); main.go defaults to 8080 if unset.
EXPOSE 8080
ENTRYPOINT ["/score"]
