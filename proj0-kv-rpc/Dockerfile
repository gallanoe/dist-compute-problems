FROM golang:1.25 as builder

# TODO: Add model build step somewhere?

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code including subdirectories
COPY . .

# Build the server binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /kvrpc-server ./server

# Run the unit tests
FROM builder AS run-test-stage
RUN go test -v ./server

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage
WORKDIR /
COPY --from=builder /kvrpc-server /kvrpc-server
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/kvrpc-server"]