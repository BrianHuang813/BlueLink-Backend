FROM golang:1.25-alpine

RUN apk add --no-cache git build-base

# 安裝新版 Air
RUN go install github.com/air-verse/air@latest

WORKDIR /app

# 複製 go.mod 生成 go.sum
COPY go.mod ./
RUN go mod tidy

# 複製程式碼
COPY . .

EXPOSE 8080
CMD ["air", "-c", ".air.toml"]
