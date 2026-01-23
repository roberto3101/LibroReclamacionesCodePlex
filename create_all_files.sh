#!/bin/bash

# Script para crear todos los archivos del proyecto

# ============= BACKEND FILES =============

# package.json backend
cat > backend/package.json << 'EOF'
{
  "name": "libro-reclamaciones-backend",
  "version": "1.0.0",
  "type": "module",
  "description": "Backend API para Libro de Reclamaciones CODEPLEX con CockroachDB",
  "main": "server.js",
  "scripts": {
    "start": "node server.js",
    "dev": "nodemon server.js",
    "db:setup": "cockroach sql --url=\"$DATABASE_URL\" -f schema.sql"
  },
  "dependencies": {
    "express": "^4.18.2",
    "cors": "^2.8.5",
    "pg": "^8.11.3",
    "nodemailer": "^6.9.7",
    "dotenv": "^16.3.1"
  },
  "devDependencies": {
    "nodemon": "^3.0.2"
  },
  "engines": {
    "node": ">=18.0.0"
  }
}
EOF

# .env.example backend
cat > backend/.env.example << 'EOF'
# =============================================================================
# CONFIGURACIÓN DEL BACKEND - LIBRO DE RECLAMACIONES
# Copiar este archivo a .env y completar con tus credenciales
# =============================================================================

# Puerto del servidor
PORT=3000

# Modo de ejecución
NODE_ENV=development

# =============================================================================
# BASE DE DATOS COCKROACHDB
# =============================================================================

# CockroachDB Serverless (RECOMENDADO - Free tier: 10 GB, 250M RUs/mes)
# 1. Crear cuenta en https://cockroachlabs.cloud/signup
# 2. Create cluster → Serverless → Region: us-east-1
# 3. Copiar connection string

DATABASE_URL=postgresql://usuario:password@host.cockroachlabs.cloud:26257/defaultdb?sslmode=verify-full

# Ejemplo local (desarrollo):
# DATABASE_URL=postgresql://root@localhost:26257/defaultdb?sslmode=disable

# =============================================================================
# CONFIGURACIÓN DE EMAIL (Nodemailer)
# =============================================================================

# Opción 1: Gmail (Requiere "App Password")
# Crear App Password: https://myaccount.google.com/apppasswords
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=tu_email@gmail.com
SMTP_PASS=tu_app_password_de_16_caracteres

# Email remitente
SMTP_FROM=libro.reclamaciones@codeplex.pe

# Email de soporte (recibe notificaciones)
EMAIL_SOPORTE=soporte@codeplex.pe

# =============================================================================
# CONFIGURACIÓN ADICIONAL
# =============================================================================

# URL del frontend (para CORS)
FRONTEND_URL=https://codeplex.pe

# Secreto para tokens (opcional)
JWT_SECRET=tu_secreto_super_seguro_cambiar_en_produccion
EOF

# ============= FRONTEND FILES =============

# package.json frontend
cat > frontend/package.json << 'EOF'
{
  "name": "libro-reclamaciones-frontend",
  "type": "module",
  "version": "1.0.0",
  "scripts": {
    "dev": "astro dev",
    "start": "astro dev",
    "build": "astro check && astro build",
    "preview": "astro preview",
    "astro": "astro"
  },
  "dependencies": {
    "astro": "^4.16.18",
    "@astrojs/tailwind": "^5.1.2",
    "tailwindcss": "^3.4.17"
  },
  "devDependencies": {
    "@astrojs/check": "^0.9.4",
    "typescript": "^5.7.3"
  }
}
EOF

# astro.config.mjs
cat > frontend/astro.config.mjs << 'EOF'
import { defineConfig } from 'astro/config';
import tailwind from '@astrojs/tailwind';

export default defineConfig({
  integrations: [tailwind()],
  output: 'static',
  site: 'https://codeplex.pe'
});
EOF

# tailwind.config.mjs
cat > frontend/tailwind.config.mjs << 'EOF'
/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{astro,html,js,jsx,md,mdx,svelte,ts,tsx,vue}'],
  theme: {
    extend: {
      colors: {
        'codeplex-blue': '#1e40af',
        'codeplex-dark': '#1f2937',
      },
    },
  },
  plugins: [],
}
EOF

# .env.example frontend
cat > frontend/.env.example << 'EOF'
# URL de la API Backend
PUBLIC_API_URL=http://localhost:3000/api

# En producción (Railway/Vercel):
# PUBLIC_API_URL=https://tu-backend.railway.app/api
EOF

# tsconfig.json
cat > frontend/tsconfig.json << 'EOF'
{
  "extends": "astro/tsconfigs/strict"
}
EOF

echo "✅ Todos los archivos de configuración creados"
echo "Ahora ejecutando el script..."
