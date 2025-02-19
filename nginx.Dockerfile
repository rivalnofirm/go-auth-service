FROM nginx:latest

# Menyalin file konfigurasi nginx
COPY ./nginx/nginx.conf /etc/nginx/nginx.conf
COPY ./nginx/default.conf /etc/nginx/conf.d/default.conf