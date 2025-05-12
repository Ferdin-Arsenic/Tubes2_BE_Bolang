# Gunakan image Golang versi terbaru
FROM golang:1.24.2

# Set working directory di dalam container
WORKDIR /app

# Salin semua file dari folder src ke dalam container
COPY src/ .

# Install dependencies Go dan persiapkan mod
RUN go mod tidy

# Build aplikasi
RUN go build -o app .

# Expose port yang digunakan oleh aplikasi (contoh: 8080)
EXPOSE 8080

# Jalankan aplikasi yang sudah dibuild
CMD ["./app"]
