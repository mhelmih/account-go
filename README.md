# Account Service

Account Service adalah layanan HTTP server menggunakan REST API yang memungkinkan pengguna untuk melakukan operasi terkait akun seperti registrasi, menabung, menarik dana, dan memeriksa saldo nasabah. Layanan ini dibangun menggunakan bahasa pemrograman Go dengan framework Echo dan menggunakan PostgreSQL sebagai database.

## Fitur

1. **Registrasi Nasabah Baru**

   - Endpoint: `/daftar`
   - Method: `POST`
   - Payload JSON: `{"nama": "John Doe", "nik": "1234567890", "no_hp": "081234567890"}`
   - Response: `{"no_rekening": "1234567890"}`

2. **Menabung**

   - Endpoint: `/tabung`
   - Method: `POST`
   - Payload JSON: `{"no_rekening": "1234567890", "nominal": 100000}`
   - Response: `{"saldo": 200000}`

3. **Menarik Dana**

   - Endpoint: `/tarik`
   - Method: `POST`
   - Payload JSON: `{"no_rekening": "1234567890", "nominal": 50000}`
   - Response: `{"saldo": 150000}`

4. **Memeriksa Saldo**
   - Endpoint: `/saldo/{no_rekening}`
   - Method: `GET`
   - Response: `{"saldo": 150000}`

## Prasyarat

- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [Go 1.21.1](https://golang.org/dl/)

## Cara Menjalankan

### Langkah 1: Clone Repository

Clone repository ini ke mesin lokal Anda:

```sh
git clone https://github.com/username/account-service.git
cd account-service
```

### Langkah 2: Membuat File .env

Buat file .env di root directory proyek dan isi dengan konfigurasi berikut:

```env
POSTGRES_USER=account_admin
POSTGRES_PASSWORD=password123
POSTGRES_DB=account
DB_HOST=db
```

### Langkah 3: Menjalankan dengan Docker Compose

Jalankan perintah berikut untuk membangun dan menjalankan layanan menggunakan Docker Compose:

```sh
docker-compose up --build
```

Perintah ini akan:

Membuild image Docker untuk layanan aplikasi.
Menjalankan container untuk layanan aplikasi dan database PostgreSQL.

### Langkah 4: Menjalankan dengan Flag untuk Mengatur Host dan Port

Anda juga dapat menjalankan aplikasi secara lokal dengan menggunakan flag untuk mengatur host dan port. Berikut adalah contoh menjalankan aplikasi dengan flag:

```sh
go run main.go --host 0.0.0.0 --port 1323
```

### Langkah 5: Mengakses Layanan

Setelah semua layanan berjalan, Anda dapat mengakses endpoint API melalui http://localhost:1323.
