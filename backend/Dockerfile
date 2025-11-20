FROM golang:alpine AS build

# Install the ca-certificate package
RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates && apk add --no-cache bash

# Move to working directory /build
WORKDIR /build

# Copy the go.mod and go.sum files to the /build directory
COPY go.mod go.sum ./

# Install dependencies
RUN go mod download && go mod verify

# Copy the entire source code into the container
COPY . .

# Build the application
# Turn off CGO to ensure static binaries
RUN CGO_ENABLED=0 go build -o app

# Production stage
# =============================================================================
# Create a production stage to run the application binary
FROM scratch

ARG UID=1000
ARG GID=1000
 
# # Copy binary from builder stage
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /build/app /app
COPY --from=build /build/sqlc/migrations /migrations

# # Start the application
CMD ["/app"]