FROM golang:1.21-alpine AS builder
WORKDIR /app

# نسخ ملف التعريف وتنزل التبعيات وتوليد البصمات تلقائياً
COPY go.mod ./
RUN go env -w GO111MODULE=on
RUN go get gopkg.in/telebot.v3

# نسخ باقي الكود وبناء البرنامج بدون التحقق الحارم من go.sum
COPY . .
RUN go build -o bot main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/bot .
CMD ["./bot"]
