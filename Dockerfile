# Tahap 1: Build Stage
FROM golang:1.21-alpine AS builder

LABEL maintener="rivalnofirm"

# Menginstal git (diperlukan untuk go mod jika menggunakan module)
RUN apk add --no-cache git

# Set environment variables
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Membuat direktori kerja
WORKDIR /app

# Menyalin go.mod dan go.sum untuk mendownload dependensi
COPY go.mod go.sum ./

# Mengunduh semua dependensi yang diperlukan
RUN go mod download

# Menyalin seluruh kode sumber aplikasi ke dalam container
COPY . .

# Membuild aplikasi
RUN go build -o app ./main.go

# Tahap 2: Runtime Stage
FROM alpine:latest

# Mengatur direktori kerja di dalam container
WORKDIR /root/

# Menyalin file binary yang telah dibuild dari tahap builder
COPY --from=builder /app/app .

# Menentukan command untuk menjalankan aplikasi
CMD ["./app"]