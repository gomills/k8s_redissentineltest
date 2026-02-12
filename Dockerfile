# Stage 1: Building the image, or, in this case, the binary, which is setting up ready for container
FROM golang:1.25-alpine AS builder

# set environment variables for build
ENV CGO_ENABLED=0

# create working directory in the building engine
WORKDIR /build

# Cache dependencies. 
# Don't copy all source for smarter caching of dependencies.
COPY go.mod go.sum ./
RUN go mod download

# now, yes, copy all source and build
COPY . .
RUN go build -o /app/redistestbin main.go

# Stage 2: Running the binary
FROM alpine:3.21 AS final

# create new directory in the container
WORKDIR /app

# copy only the binary from the builder source files
COPY --from=builder /app/redistestbin .

# create user to run binary
RUN addgroup -S app && adduser -S app -G app && \
    chown app:app /app/redistestbin

USER app

CMD ["/app/redistestbin"]