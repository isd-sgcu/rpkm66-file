# Base Image
FROM golang:1.20.5-bullseye as base

# Working directory
WORKDIR /app

# Setup credential
RUN git config --global url.ssh://git@github.com/.insteadOf https://github.com/ && mkdir /root/.ssh && chmod 700 /root/.ssh && ssh-keyscan github.com >> /root/.ssh/known_hosts

ENV GOPRIVATE=github.com/isd-sgcu/*

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN --mount=type=secret,id=sshKey,target=/root/.ssh/id_rsa,required=true go mod download

# Copy the source code
COPY . .

# Build the application
RUN --mount=type=secret,id=sshKey,target=/root/.ssh/id_rsa,required=true go build -o server ./cmd/main.go
# Create master image
FROM alpine AS master

# Working directory
WORKDIR /app

# Copy execute file
COPY --from=base /internal/server ./

# Set ENV to production
ENV GO_ENV production

# Expose port 3003
EXPOSE 3003

# Run the application
CMD ["./server"]