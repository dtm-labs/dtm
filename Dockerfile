FROM daocloud.io/atsctoo/golang:1.15
WORKDIR /app/dtm
COPY . .
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
RUN go build app/main.go
EXPOSE 8080
CMD [ "/app/dtm/main", "dtmsvr" ]
