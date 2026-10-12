FROM golang:1.25.5-bookworm@sha256:d9132cce84391efab786495288756d60e1da215b1f94e87860aeefc3d4c45b6d
RUN apt-get update && apt-get install -y --no-install-recommends pkg-config libhyperscan-dev=5.4.0-2 && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=1
RUN go build -o /usr/local/bin/baseline ./cmd/baseline && go build -tags hyperscan -o /usr/local/bin/hyperscan ./cmd/hyperscan
WORKDIR /work
