FROM golang:1.15
WORKDIR /app/dtm
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.io,direct
EXPOSE 8080
CMD [ "/bin/bash", "-c", "go build app/main.go && /app/dtm/main dev"]
# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Builder
FROM golang:1.15 as builder

WORKDIR /src

COPY . .

RUN go env -w GO111MODULE=on \
    && go env -w GOPROXY=https://goproxy.io,direct \
    && go mod download \ 
    && go build -o dtm app/main.go

# Runtimer 
FROM debian:unstable-slim

WORKDIR /svr

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /src/dtm .
COPY /dtmsvr/dtmsvr.mysql.sql ./dtmsvr/dtmsvr.mysql.sql
COPY /examples/examples.mysql.sql ./examples/examples.mysql.sql
COPY /dtmcli/barrier.mysql.sql ./dtmcli/barrier.mysql.sql

# Expose port 80 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["sh", "-c", "./dtm dev"]