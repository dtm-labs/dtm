FROM daocloud.io/golang:1.15
WORKDIR /app/dtm
COPY . .
RUN go build app/main.go
EXPOSE 8080
CMD [ "/app/dtm/main", "dtmsvr" ]
