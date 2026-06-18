# Voxlab Backend Deployment Guide (VPS)

## Overview

Voxlab Backend is fully Dockerized. All services run in containers:
- Backend API (Go + Gin)
- PostgreSQL (with pg_cron)
- Redis
- MinIO
- pgAdmin
- LiveKit
- Python Analyzer

## Prerequisites

- Ubuntu 22.04 LTS or later
- Docker 24.0+
- Docker Compose v2
- Domain name (optional, for HTTPS)
- SSL certificates (optional, via Let's Encrypt)

## 1. VPS Setup

### Install Docker
```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com | sh

# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Install Docker Compose plugin
sudo apt install docker-compose-plugin -y
```

### Verify installation
```bash
docker --version
docker compose version
```

## 2. Clone Repository

```bash
cd /home/andres/projects
git clone https://github.com/D-Kale/Voxlab-Backend.git voxlab-backend
cd voxlab-backend
```

## 3. Environment Configuration

```bash
cp .env.example .env
nano .env
```

### Required Environment Variables

See `docs/ENV_VARS.md` for complete list.

**Minimum required for production:**
```env
APP_ENV=production
APP_PORT=3000
DATABASE_URL=postgres://postgres:postgres@postgres:5432/voxlab?sslmode=disable
REDIS_URL=redis:6379
JWT_SECRET=change_this_to_a_long_random_string
MINIO_ENDPOINT=minio:9000
MINIO_PUBLIC_ENDPOINT=https://media.voxlab.app
MINIO_ACCESS_KEY=change_this
MINIO_SECRET_KEY=change_this
MINIO_BUCKET=voxlab-media
ANALYZER_URL=http://analyzer:8000
LIVEKIT_HOST=http://livekit:7880
```

## 4. Start Services

```bash
docker compose up -d --build
```

### Check status
```bash
docker compose ps
```

Expected output:
```
NAME              STATUS                  PORTS
voxlab-backend    Up (healthy)             0.0.0.0:3000->3000/tcp
voxlab-postgres   Up (healthy)             0.0.0.0:5433->5432/tcp
voxlab-redis      Up (healthy)             0.0.0.0:6380->6379/tcp
voxlab-minio      Up (healthy)             0.0.0.0:9010->9000/tcp
voxlab-pgadmin    Up                       0.0.0.0:5050->80/tcp
voxlab-livekit    Up                       0.0.0.0:7880->7880/tcp
voxlab-analyzer   Up                       0.0.0.0:8001->8000/tcp
```

## 5. Verify Deployment

### Health check
```bash
curl http://localhost:3000/api/v1/health
```

Expected response:
```json
{
  "services": {
    "postgres": "ok",
    "redis": "ok"
  },
  "status": "ok",
  "timestamp": "2026-06-18T...",
  "version": "1.0.0"
}
```

### Swagger UI
```bash
http://your-domain.com/swagger/index.html
```

### Spanish docs
```bash
http://your-domain.com/docs/es
```

## 6. Configure Firewall

```bash
# Allow SSH
sudo ufw allow 22

# Allow HTTP/HTTPS
sudo ufw allow 80
sudo ufw allow 443

# Allow API port (if not using reverse proxy)
sudo ufw allow 3000

# Enable firewall
sudo ufw enable
```

## 7. Configure Nginx (Optional, recommended for HTTPS)

### Install Nginx
```bash
sudo apt install nginx -y
```

### Create Nginx config
```bash
sudo nano /etc/nginx/sites-available/voxlab
```

```nginx
server {
    listen 80;
    server_name api.voxlab.app media.voxlab.app;

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /socket {
        proxy_pass http://localhost:7880;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }
}
```

### Enable site
```bash
sudo ln -s /etc/nginx/sites-available/voxlab /etc/nginx/sites-enabled/voxlab
sudo nginx -t
sudo systemctl reload nginx
```

## 8. SSL with Let's Encrypt

```bash
sudo apt install certbot python3-certbot-nginx -y
sudo certbot --nginx -d api.voxlab.app -d media.voxlab.app
```

## 9. Backup Strategy

### Database backup
```bash
docker compose exec -T postgres pg_dump -U postgres voxlab > backup_$(date +%Y%m%d).sql
```

### MinIO backup
```bash
mc alias set local http://localhost:9010 minioadmin minioadmin
mc mirror --overwrite local/voxlab-media ./backup/minio
```

## 10. Monitoring

### Logs
```bash
docker compose logs -f backend
docker compose logs -f postgres
docker compose logs -f minio
```

### Resource usage
```bash
docker stats
```

### Restart services
```bash
docker compose restart backend
```

## 11. Update Deployment

```bash
# Pull latest changes
git pull origin master

# Rebuild and restart
docker compose up -d --build
```

## 12. Troubleshooting

### Port conflicts
```bash
sudo lsof -i :3000
sudo fuser -k 3000/tcp
```

### Database volume issue (PG15 → PG16)
```bash
docker compose down -v
docker compose up -d --build
```

### MinIO bucket not accessible
```bash
docker compose logs minio
```

### Analyzer unavailable (502)
```bash
docker compose logs analyzer
docker compose restart analyzer
```

## 13. Security Checklist

- [ ] Change all default passwords in `.env`
- [ ] Use strong `JWT_SECRET` (min 32 chars)
- [ ] Enable HTTPS with Let's Encrypt
- [ ] Configure firewall (ufw)
- [ ] Restrict pgAdmin access (VPN or IP whitelist)
- [ ] Use non-root container user (backend runs as appuser)
- [ ] Rotate MinIO credentials periodically
- [ ] Set up automated backups