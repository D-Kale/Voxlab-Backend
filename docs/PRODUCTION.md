# Voxlab — Guía de Producción

## 📦 Stack Completo

| Servicio   | Contenedor        | Puerto Interno | Puerto Host | Rol                |
|------------|-------------------|----------------|-------------|--------------------|
| nginx      | voxlab-nginx      | 80 / 443       | 80 / 443    | Reverse proxy + SSL|
| frontend   | voxlab-frontend   | 3001           | —           | Astro SSR          |
| backend    | voxlab-backend    | 3000           | 3000        | Go API             |
| analyzer   | voxlab-analyzer   | 8000           | 8001        | Python writing analyzer|
| postgres   | voxlab-postgres   | 5432           | 5433        | Base de datos      |
| redis      | voxlab-redis      | 6379           | 6380        | Cache              |
| minio      | voxlab-minio      | 9000 / 9001    | 9010 / 9011 | S3-compatible storage|
| livekit    | voxlab-livekit    | 7880 / 7881    | 7883 / 7884 | WebRTC             |
| pgadmin    | voxlab-pgadmin    | 80             | 5050        | DB admin (dev)     |

---

## 🚀 Deploy Inicial

```bash
# En el VPS Hostinger (8GB RAM / 100GB SSD)
ssh root@IP_DEL_VPS

# Instalar Docker + compose
apt update && apt install -y docker.io docker-compose-v2

# Clonar repos
git clone git@github.com:D-Kale/Voxlab-Backend.git /opt/voxlab/backend
git clone <frontend-repo> /opt/voxlab/Frontend

# Crear estructura
mkdir -p /opt/voxlab/backend/nginx/ssl /opt/voxlab/backend/backups

# Variables de entorno (producción)
cat > /opt/voxlab/backend/.env << EOF
JWT_SECRET=$(openssl rand -hex 32)
MINIO_ACCESS_KEY=$(openssl rand -hex 16)
MINIO_SECRET_KEY=$(openssl rand -hex 32)
LIVEKIT_API_KEY=prodkey
LIVEKIT_API_SECRET=$(openssl rand -hex 32)
EOF

# Levantar todo
cd /opt/voxlab/backend
docker compose up -d --build

# Verificar
docker compose ps
docker compose logs -f --tail=50
```

---

## 🌐 Conectar Dominio (Hostinger)

Hostinger permite gestionar dominio + VPS desde el mismo panel.

### Paso 1 — Reclamar el dominio

1. Ir a **Hostinger hPanel** → **Dominios**
2. Si ya lo compraste: aparece en "Mis Dominios"
3. Si no: usa el buscador para comprar uno nuevo (~$10-15/año)

### Paso 2 — Apuntar DNS al VPS

En hPanel → **Dominios** → **DNS Zone**:

| Tipo | Nombre | Valor                    | TTL  |
|------|--------|--------------------------|------|
| A    | @      | `IP_PÚBLICA_DEL_VPS`     | 3600 |
| A    | www    | `IP_PÚBLICA_DEL_VPS`     | 3600 |

> La IP del VPS la ves en hPanel → **VPS** → **IP Address**

### Paso 3 — Configurar nginx con tu dominio

Editar `nginx/nginx.conf` y **descomentar el bloque HTTPS** (`server { listen 443 ssl ... }`):
- Cambiar `server_name` por tu dominio (ej: `voxlab.com www.voxlab.com`)
- El bloque HTTP (puerto 80) hará redirect automático 301 a HTTPS

### Paso 4 — Obtener SSL (Let's Encrypt)

Con la stack ya levantada:

```bash
# Certbot standalone (temporalmente detiene nginx)
cd /opt/voxlab/backend

docker compose stop nginx

docker run --rm -p 80:80 \
  -v $(pwd)/nginx/ssl:/etc/letsencrypt \
  certbot/certbot certonly --standalone \
  -d tudominio.com -d www.tudominio.com \
  --agree-tos --no-eff-email -m tu@email.com

# Copiar certificados a la ubicación que espera nginx
cp -L /etc/letsencrypt/live/tudominio.com/fullchain.pem nginx/ssl/
cp -L /etc/letsencrypt/live/tudominio.com/privkey.pem nginx/ssl/

docker compose start nginx
```

### Paso 5 — Auto-renewal (cron)

```bash
crontab -e
# Agregar:
0 4 * * * cd /opt/voxlab/backend && docker compose stop nginx && docker run --rm -p 80:80 -v $(pwd)/nginx/ssl:/etc/letsencrypt certbot/certbot renew --quiet && cp -L /etc/letsencrypt/live/tudominio.com/fullchain.pem nginx/ssl/ && cp -L /etc/letsencrypt/live/tudominio.com/privkey.pem nginx/ssl/ && docker compose start nginx
```

### Paso 6 — Verificar

```bash
curl -I https://tudominio.com           # → 200 OK + HSTS header
curl https://tudominio.com/api/v1/health # → backend responde
```

---

## 🔒 Seguridad

### Lo que ya está configurado (nginx)

- **Rate limiting**: 20 req/s a API, 10 req/s a frontend
- **Límite de conexiones**: 10 por IP a API, 5 a frontend
- **Security headers**: X-Frame-Options, X-Content-Type-Options, XSS Protection, Referrer-Policy, Permissions-Policy
- **Server tokens**: off (no revela versión de nginx)
- **Tamaño máximo de upload**: 50MB (configurable)

### Recomendaciones post-deploy

- [ ] Cambiar todas las contraseñas por defecto en `.env`
- [ ] Exponer **solo puertos 80/443** (nginx), cerrar los demás en firewall
- [ ] Configurar backups automáticos de PostgreSQL
- [ ] Monitoreo con `docker compose logs` + alertas

### Firewall (UFW)

```bash
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp        # SSH
ufw allow 80/tcp        # HTTP
ufw allow 443/tcp       # HTTPS
ufw enable
```

---

## 💾 Backup PostgreSQL

El backup se hace con un contenedor que ejecuta `pg_dump` diariamente.

Para activarlo, descomentar en `docker-compose.yml` el servicio `postgres-backup`:

```yaml
  postgres-backup:
    image: postgres:16-alpine
    container_name: voxlab-postgres-backup
    volumes:
      - ./backups:/backups
    environment:
      - PGHOST=postgres
      - PGUSER=postgres
      - PGPASSWORD=postgres
      - PGDATABASE=voxlab
    command: >
      sh -c "while true; do
        pg_dump -h postgres -U postgres voxlab | gzip > /backups/voxlab_\$(date +%Y%m%d_%H%M%S).sql.gz;
        find /backups -name '*.sql.gz' -mtime +7 -delete;
        sleep 86400;
      done"
    depends_on:
      - postgres
    networks:
      - voxlab
    restart: unless-stopped
```

Los backups quedan en `backend/backups/` con retención de 7 días.

### Restaurar un backup

```bash
gunzip -c backups/voxlab_20250625_020000.sql.gz | docker exec -i voxlab-postgres psql -U postgres voxlab
```

---

## 📊 Recursos (VPS 8GB RAM)

| Servicio   | RAM     | CPU | Disco      |
|------------|---------|-----|------------|
| postgres   | ~1.5 GB | 0.5 | variable   |
| redis      | ~128 MB | 0.1 | ~100 MB    |
| minio      | ~512 MB | 0.2 | según assets|
| livekit    | ~512 MB | 0.3 | —          |
| analyzer   | ~512 MB | 0.5 | —          |
| backend    | ~512 MB | 0.5 | —          |
| frontend   | ~384 MB | 0.2 | —          |
| nginx      | ~128 MB | 0.1 | —          |
| **Total**  | **~4.2 GB** | **2.8 cores** | ~100 GB |

---

## 🐛 Troubleshooting

**El frontend no arranca**
```bash
docker compose logs frontend
# Verificar PUBLIC_API_URL apunte a http://backend:3000/api/v1
```

**Error de conexión a DB**
```bash
docker compose logs postgres
docker compose exec postgres pg_isready -U postgres
```

**nginx no levanta**
```bash
docker compose logs nginx
# Verificar sintaxis del config:
docker compose exec nginx nginx -t
```

**Certificados SSL vencidos**
```bash
# Renew manual:
docker compose stop nginx && docker run --rm -p 80:80 -v $(pwd)/nginx/ssl:/etc/letsencrypt certbot/certbot renew && docker compose start nginx
```
