#!/bin/bash

echo "🚀 Generando proyecto completo..."

# Backend package.json
cat > backend/package.json << 'EOF'
{
  "name": "libro-reclamaciones-backend",
  "version": "1.0.0",
  "type": "module",
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
  }
}
EOF

# Backend .env.example
cat > backend/.env.example << 'EOF'
PORT=3000
NODE_ENV=development

# CockroachDB Serverless
DATABASE_URL=postgresql://usuario:password@host.cockroachlabs.cloud:26257/defaultdb?sslmode=verify-full

# Email (Gmail con App Password)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=tu_email@gmail.com
SMTP_PASS=tu_app_password
SMTP_FROM=libro.reclamaciones@codeplex.pe
EMAIL_SOPORTE=soporte@codeplex.pe

FRONTEND_URL=https://codeplex.pe
EOF

# Frontend package.json
mkdir -p frontend/src/{pages,layouts,components}
cat > frontend/package.json << 'EOF'
{
  "name": "libro-reclamaciones-frontend",
  "type": "module",
  "version": "1.0.0",
  "scripts": {
    "dev": "astro dev",
    "build": "astro check && astro build",
    "preview": "astro preview"
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

# tsconfig.json
cat > frontend/tsconfig.json << 'EOF'
{
  "extends": "astro/tsconfigs/strict"
}
EOF

# .env.example
cat > frontend/.env.example << 'EOF'
PUBLIC_API_URL=http://localhost:3000/api
EOF

# Layout.astro
cat > frontend/src/layouts/Layout.astro << 'EOF'
---
interface Props {
  title: string;
}
const { title } = Astro.props;
---
<!doctype html>
<html lang="es">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>{title} | CODEPLEX</title>
  </head>
  <body class="bg-gray-50">
    <slot />
  </body>
</html>
<style is:global>
  html { font-family: system-ui, sans-serif; scroll-behavior: smooth; }
</style>
EOF

# .gitignore
cat > .gitignore << 'EOF'
node_modules/
.env
dist/
.DS_Store
*.log
EOF

echo "✅ Proyecto base generado"
