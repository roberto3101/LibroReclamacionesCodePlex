# 🚀 INSTRUCCIONES RÁPIDAS DE INSTALACIÓN

## ✅ LO QUE TIENES

Un sistema completo de Libro de Reclamaciones con:
- ✅ **CockroachDB** (base de datos distribuida)
- ✅ **Formularios dinámicos** (RECLAMO/QUEJA como botones clickeables)
- ✅ **Firma digital obligatoria** (canvas con signature_pad)
- ✅ **Backend Node.js/Express**
- ✅ **Frontend Astro + Tailwind**

## 📦 INSTALACIÓN EN 10 MINUTOS

### 1. Backend (5 min)

```bash
cd backend

# Instalar
npm install

# Configurar CockroachDB
# 1. Ir a: https://cockroachlabs.cloud/signup
# 2. Create Cluster → Serverless (GRATIS)
# 3. Copiar connection string

# Crear .env
cp .env.example .env
# Editar .env y pegar tu DATABASE_URL

# Crear tablas
npm run db:setup

# Iniciar
npm run dev
```

### 2. Frontend (5 min)

```bash
cd frontend

# Instalar
npm install

# Crear .env
cp .env.example .env

# Iniciar
npm run dev
```

## 🎯 CARACTERÍSTICAS IMPLEMENTADAS

### UX Mejorada
1. **Pantalla inicial**: 2 botones grandes clickeables
   - 🔴 RECLAMO (producto/servicio)
   - 🟡 QUEJA (atención al cliente)

2. **Formulario dinámico**: Cambia según selección
   - RECLAMO: muestra campos de producto, monto, etc.
   - QUEJA: muestra campos de área, personal, situación

3. **Firma digital**: Al aceptar términos aparece canvas
   - Obligatoria para enviar
   - Se guarda como base64 en CockroachDB

### Backend Robusto
- ✅ CockroachDB Serverless (escalable, distribuido)
- ✅ Generación automática de códigos únicos
- ✅ Envío de emails automático (consumidor + soporte)
- ✅ API RESTful completa
- ✅ Dashboard de estadísticas

## 📧 Configuración de Gmail

Para enviar emails:

1. Ir a: https://myaccount.google.com/apppasswords
2. Crear "App Password" para Mail
3. Copiar contraseña de 16 caracteres
4. En `backend/.env`:
   ```
   SMTP_USER=tu_email@gmail.com
   SMTP_PASS=xxxx xxxx xxxx xxxx
   ```

## 🚀 Deploy a Producción

### Backend → Railway
```bash
npm install -g @railway/cli
cd backend
railway login
railway init
railway up
```

### Frontend → Vercel
```bash
npm install -g vercel
cd frontend
vercel
```

## 📱 Probar Localmente

1. Backend: http://localhost:3000/api/health
2. Frontend: http://localhost:4321

## ✅ Cumplimiento Legal

- ✅ Ley N° 29571
- ✅ D.S. 011-2011-PCM  
- ✅ D.S. 006-2014-PCM
- ✅ Ley N° 27269 (Firma Digital)

## 💡 Soporte

CODEPLEX: soporte@codeplex.pe | +51 936343607
INDECOPI: 224-7777 (Lima) | 0-800-4-4040 (Provincias)
