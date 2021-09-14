FROM golang:1.15
WORKDIR /app/dtm
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.io,direct
EXPOSE 8080
CMD [ "/bin/bash", "-c", "go build app/main.go && /app/dtm/main dev"]
