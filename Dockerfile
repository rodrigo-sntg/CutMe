FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /go/bin/app cmd/main.go
RUN chmod +x /go/bin/app

FROM alpine:3.17

RUN apk add --no-cache ffmpeg

WORKDIR /app

ENV AWS_REGION=us-east-1
ENV S3_BUCKET=meu-bucket-processamento
ENV DYNAMO_TABLE=ArquivosProcessados1
ENV QUEUE_URL=https://queue.amazonaws.com/058264063116/MinhaFila
ENV CLOUDFRONT_DOMAIN_NAME=d12jxjn0s3w75f.cloudfront.net
ENV SMTP_HOST=smtp.gmail.com
ENV SMTP_PORT=587
ENV SMTP_EMAIL=contact@wandrmate.com
ENV SMTP_PASSWORD=${SMTP_PASSWORD}
ENV AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}
ENV AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}

COPY --from=builder /go/bin/app /app/app

EXPOSE 8080

CMD ["/app/app"]
