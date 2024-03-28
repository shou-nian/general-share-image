FROM golang:1.21.4

WORKDIR /app

COPY go.mod .
COPY go.sum .

# 下载依赖包
RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 8080

CMD ["./main"]