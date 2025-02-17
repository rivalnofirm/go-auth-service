# go-auth-service

## Overview
Repository ini berisikan service API dan worker yang merujuk pada sekumpulan endpoint yang digunakan untuk menangani proses autentikasi dan otorisasi pengguna dalam suatu sistem.

## Teknologi
Beberapa teknologi yang digunakan :
- [Golang](https://go.dev/)
- [PostgreSQL](https://www.postgresql.org/)
- [NATS](https://nats.io/)
- [DOCKER](https://www.docker.com/)
- [NGINX](https://nginx.org/)

## Instalasi


## API
berikut endpoint yang ada didalam repository ini 

| Endpoint                                 | Method | Description                                                                |
|------------------------------------------|--------|----------------------------------------------------------------------------|
| `/api/auth/register`                     | `POST` | Endpoint untuk mendaftarkan akun baru.                                     |
| `/api/auth/login`                        | `POST` | Endpoint untuk masuk ke sistem dan mendapatkan access token.               |
| `/api/auth/me`                           | `GET`  | Mengambil informasi akun pengguna yang sedang login.                       |
| `/api/auth/refresh-token`                | `GET`  | Memperbarui access token yang sudah kedaluwarsa menggunakan refresh token. |
| `/api/auth/logout`                       | `GET`  | Logout pengguna dan menghapus sesi atau refresh token.                     |
| `/api/auth/revoke-token/{email-encrypt}` | `GET`  | Menonaktifkan atau mencabut token akses berdasarkan email yang dienkripsi. |
| `/api/auth/update-profile`               | `PUT`  | Memperbarui informasi profil pengguna.                                     |
| `/api/auth/update-profile-picture`       | `PUT`  | Mengunggah atau memperbarui foto profil pengguna.                          |


## Example Request

- register
    ```
  curl --location 'http://127.0.0.1:8080/api/user/register' \
        --header 'Content-Type: application/json' \
        --data-raw '{
        "first_name": "abdul",
        "last_name": "gofur",
        "email": "example@gmail.com",
        "password": "test12345"
        }
        '
  ```
  
- login
  ```
  curl --location 'http://127.0.0.1:8080/api/user/login' \
    --header 'Content-Type: application/json' \
    --data-raw '{
    "email" : "example@gmail.com",
    "password": "test12345"
    }'
  ```
  
- me
  ```
  curl --location 'http://127.0.0.1:8080/api/user/me' \
  --header 'Authorization: (gunakan access_token generate dari login) '
  ```
  
- refresh token
  ```
  curl --location 'http://127.0.0.1:8080/api/user/refresh-token' \
  --header 'Authorization: (gunakan refresh_token generate dari login)'
  ```
  
- logout
  ```
  curl --location 'http://127.0.0.1:8080/api/user/logout' \
  --header 'Authorization: (gunakan access_token generate dari login)'
  ```
  
- revoke token
  ```
  curl --location 'http://127.0.0.1:8080/api/user/revoke-token/(menggunakan email yang di generate otomatis ketika user login dengan device yang berbeda)'
  ```
  
- update profile
  ```
  curl --location --request PUT 'http://127.0.0.1:8080/api/user/update-profile' \
    --header 'Authorization: (gunakan access_token generate dari login)' \
    --header 'Content-Type: application/json' \
    --data '{
    "first_name": "kulu",
    "last_name": "kulu",
    "birth_date": "1980-01-02",
    "gender": "male"
    }
  '
  ```
  
- update profile picture
  ```
  curl --location --request PUT 'http://127.0.0.1:8080/api/user/update-profile-picture' \
  --header 'Authorization: (gunakan access_token generate dari login)' \
  --form 'profile_picture=@"/Users/rivalnofirm/Downloads/example.jpg"'
  ```

# under development




