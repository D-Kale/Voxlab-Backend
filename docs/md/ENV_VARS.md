# Environment Variables Configuration

This document lists all environment variables used by the Voxlab Backend. Use these to configure your `.env` file.

## 🚀 Core Settings

| Variable | Default | Description |
|---|---|---|
| `APP_ENV` | `development` | Application environment (`development`, `production`, `test`) |
| `APP_PORT` | `3000` | Port the Go API server listens on |

## 🗄️ Database (PostgreSQL)

| Variable | Default | Description |
|---|---|---|
| `DATABASE_URL` | `postgres://voxlab:voxlab_pass_2024@localhost:5432/voxlab_db?sslmode=disable` | Connection string for GORM (PostgreSQL 16) |

## ⚡ Cache (Redis)

| Variable | Default | Description |
|---|---|---|
| `REDIS_URL` | `localhost:6379` | Redis server address |

## 📦 Storage (MinIO / S3)

| Variable | Default | Description |
|---|---|---|
| `MINIO_ENDPOINT` | `localhost:9000` | MinIO API endpoint |
| `MINIO_PUBLIC_ENDPOINT` | `http://localhost:9000` | Public URL for accessing uploaded files |
| `MINIO_ACCESS_KEY` | `voxlab_minio_admin` | MinIO Access Key (S3 Access Key ID) |
| `MINIO_SECRET_KEY` | `voxlab_minio_pass_2024` | MinIO Secret Key (S3 Secret Access Key) |
| `MINIO_BUCKET` | `voxlab-media` | Bucket name used for storing uploads |

## 🔑 Authentication & Security

| Variable | Default | Description |
|---|---|---|
| `JWT_SECRET` | `voxlab_jwt_super_secret_key_2024_change_in_production` | Secret key used for signing JWT tokens |
| `ADMIN_EMAIL` | `admin@voxlab.com` | Email used for the automatically seeded admin user |
| `ADMIN_PASSWORD` | `admin123` | Password for the automatically seeded admin user |

## 🎙️ LiveKit (WebRTC)

| Variable | Default | Description |
|---|---|---|
| `LIVEKIT_HOST` | `http://localhost:7880` | LiveKit server endpoint |
| `LIVEKIT_API_KEY` | `devkey` | API Key for LiveKit authentication |
| `LIVEKIT_API_SECRET` | `secret` | API Secret for LiveKit authentication |

## 🤖 Python Analyzer

| Variable | Default | Description |
|---|---|---|
| `ANALYZER_URL` | `http://analyzer:8000` | Endpoint of the Python NLP analyzer microservice |

---

## 🛠️ Example `.env` File

```env
# Application
APP_ENV=development
APP_PORT=3000

# Database
DATABASE_URL=postgres://voxlab:voxlab_pass_2024@postgres:5432/voxlab?sslmode=disable

# Cache
REDIS_URL=redis:6379

# Storage
MINIO_ENDPOINT=minio:9000
MINIO_PUBLIC_ENDPOINT=http://localhost:9000
MINIO_ACCESS_KEY=voxlab_minio_admin
MINIO_SECRET_KEY=voxlab_minio_pass_2024
MINIO_BUCKET=voxlab-media

# Auth
JWT_SECRET=your_very_secret_long_random_string_here
ADMIN_EMAIL=admin@voxlab.com
ADMIN_PASSWORD=my_secure_admin_pass

# Services
ANALYZER_URL=http://analyzer:8000
LIVEKIT_HOST=http://livekit:7880
LIVEKIT_API_KEY=devkey
LIVEKIT_API_SECRET=secret
```