# Voxlab Backend

## Estructura del Proyecto

```
backend/
├── cmd/
│   └── main.go              # Punto de entrada de la aplicación
├── internal/
│   ├── config/              # Configuración y variables de entorno
│   ├── database/            # Conexiones a DB y migraciones
│   ├── handlers/            # Controladores HTTP
│   ├── middleware/          # Middleware (CORS, Auth, etc.)
│   └── models/              # Modelos GORM
├── pkg/
│   ├── response/            # Respuestas HTTP estandarizadas
│   └── storage/             # Cliente MinIO/S3
├── database/
│   └── init.sql             # Script de inicialización de DB
├── docs/                    # Documentación Swagger
├── go.mod
├── go.sum
└── Dockerfile
```

## Instalación

```bash
cd backend
go mod download
```

## Ejecución

```bash
go run cmd/main.go
```

## Documentación API

Una vez ejecutando, acceder a:
http://localhost:3000/swagger/index.html