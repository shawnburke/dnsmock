FROM golang:1.19-alpine

# Create a workspace for the app
WORKDIR /app

# fetch deps
COPY go.mod .
COPY go.sum .
RUN go mod download

# build server
COPY . .
RUN go build -o /app/dnsmock ./cmd

EXPOSE 53/udp
EXPOSE 53/tcp

ENTRYPOINT ["/app/dnsmock"]