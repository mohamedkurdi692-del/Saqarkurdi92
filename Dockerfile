FROM golang:1.21-alpine AS builder
WORKDIR /app

# نسخ ملف go.mod وتنزيل التبعيات وتوليد ملف البصمات تلقائياً
COPY go.mod ./
RUN go env -w GO111MODULE=on
RUN go get gopkg.in/telebot.v3
RUN go mod tidy

# نسخ بقية الكود وبناء تطبيق Go
COPY . .
RUN go build -o bot main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/bot .
CMD ["./bot"]
