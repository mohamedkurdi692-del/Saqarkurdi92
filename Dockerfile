FROM golang:1.21-alpine AS builder
WORKDIR /app

# نسخ ملف go.mod
COPY go.mod ./

# إجبار Go على تحميل المكتبات وتوليد go.sum تلقائياً
RUN go env -w GO111MODULE=on
RUN go get gopkg.in/telebot.v3@v3.2.1
RUN go mod tidy

# نسخ باقي الملفات وبناء التطبيق
COPY . .
RUN go build -o bot main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/bot .
CMD ["./bot"]
