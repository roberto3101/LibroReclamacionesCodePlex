// =============================================================================
// API BACKEND - LIBRO DE RECLAMACIONES CODEPLEX
// Node.js + Express + CockroachDB
// =============================================================================

import express from 'express';
import cors from 'cors';
import pkg from 'pg';
const { Pool } = pkg;
import nodemailer from 'nodemailer';
import dotenv from 'dotenv';

// Cargar .env según entorno
const envFile = process.env.NODE_ENV === 'production' ? '.env.production' : '.env';
dotenv.config({ path: envFile });

const app = express();
const PORT = process.env.PORT || 3000;

// =============================================================================
// CONFIGURACIÓN
// =============================================================================

// Middleware
app.use(cors());
app.use(express.json({ limit: '10mb' })); // Para firmas base64
app.use(express.urlencoded({ extended: true }));

// PostgreSQL/CockroachDB Pool (compatible con ambos)
const pool = new Pool({
    connectionString: process.env.DATABASE_URL,
    ssl: process.env.NODE_ENV === 'production'
        ? { rejectUnauthorized: false }
        : false
});

// Verificar conexión
pool.on('connect', () => {
    console.log('✅ Conectado a la base de datos');
});

pool.on('error', (err) => {
    console.error('❌ Error en la base de datos:', err);
});

// Nodemailer
const transporter = nodemailer.createTransport({
    host: process.env.SMTP_HOST || 'smtp.gmail.com',
    port: parseInt(process.env.SMTP_PORT || '587'),
    secure: process.env.SMTP_PORT === '465',
    auth: {
        user: process.env.SMTP_USER,
        pass: process.env.SMTP_PASS
    }
});

// =============================================================================
// FUNCIÓN: Generar código único
// =============================================================================
async function generarCodigoReclamo() {
    const año = new Date().getFullYear();

    const result = await pool.query(`
        SELECT codigo_reclamo FROM reclamos 
        WHERE codigo_reclamo LIKE $1 
        ORDER BY codigo_reclamo DESC LIMIT 1
    `, [`CODEPLEX-${año}-%`]);

    let numero = 1;
    if (result.rows.length > 0) {
        const ultimo = result.rows[0].codigo_reclamo;
        numero = parseInt(ultimo.split('-')[2]) + 1;
    }

    return `CODEPLEX-${año}-${String(numero).padStart(5, '0')}`;
}

// =============================================================================
// RUTAS DE LA API
// =============================================================================

// Health check
app.get('/api/health', async (req, res) => {
    let dbStatus = 'disconnected';
    try {
        await pool.query('SELECT 1');
        dbStatus = 'connected';
    } catch (error) {
        console.error('Health check DB error:', error.message);
    }

    res.json({
        status: 'ok',
        timestamp: new Date().toISOString(),
        database: dbStatus
    });
});

// =============================================================================
// POST /api/reclamos - Crear nuevo reclamo
// =============================================================================
app.post('/api/reclamos', async (req, res) => {
    const client = await pool.connect();

    try {
        await client.query('BEGIN');

        const {
            tipo_solicitud,
            nombre_completo,
            tipo_documento,
            numero_documento,
            telefono,
            email,
            domicilio,
            departamento,
            provincia,
            distrito,
            tipo_bien,
            monto_reclamado,
            descripcion_bien,
            area_queja,
            descripcion_situacion,
            fecha_incidente,
            detalle_reclamo,
            pedido_consumidor,
            firma_digital,
            acepta_terminos,
            acepta_copia
        } = req.body;

        // Validaciones
        if (!tipo_solicitud || !['RECLAMO', 'QUEJA'].includes(tipo_solicitud)) {
            throw new Error('Tipo de solicitud inválido');
        }

        if (!firma_digital || !firma_digital.startsWith('data:image')) {
            throw new Error('Firma digital requerida');
        }

       if (!acepta_terminos) {
            throw new Error('Debe aceptar los términos y condiciones');
        }

        // Nueva validación de Email (Test 9)
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!email || !emailRegex.test(email)) {
            throw new Error('Formato de correo electrónico inválido');
        }

       // Validación de campos mínimos y longitud (Seguridad contra ataques de saturación)
        if (!descripcion_bien || !detalle_reclamo || !pedido_consumidor) {
            throw new Error('Faltan detalles del reclamo o el pedido del consumidor');
        }

       if (detalle_reclamo.length > 3000 || pedido_consumidor.length > 2000 || nombre_completo.length > 200 || descripcion_bien.length > 600) {
            throw new Error('Uno de los campos excede el límite permitido de caracteres.');
        }

        // Generar código único
        const codigo_reclamo = await generarCodigoReclamo();

        // Insertar reclamo
        const insertQuery = `
            INSERT INTO reclamos (
                codigo_reclamo, tipo_solicitud, nombre_completo, tipo_documento, numero_documento,
                telefono, email, domicilio, departamento, provincia, distrito,
                tipo_bien, monto_reclamado, descripcion_bien, area_queja, descripcion_situacion,
                fecha_incidente, detalle_reclamo, pedido_consumidor, firma_digital,
                acepta_terminos, acepta_copia, ip_address, user_agent
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)
            RETURNING id, codigo_reclamo, fecha_registro, fecha_limite_respuesta
        `;

        const values = [
            codigo_reclamo, tipo_solicitud, nombre_completo, tipo_documento, numero_documento,
            telefono, email, domicilio || null, departamento || null, provincia || null, distrito || null,
            tipo_bien || null, monto_reclamado || 0, descripcion_bien,
            area_queja || null, descripcion_situacion || null,
            fecha_incidente, detalle_reclamo, pedido_consumidor, firma_digital,
            acepta_terminos, acepta_copia,
            req.ip, req.get('user-agent')
        ];

        const result = await client.query(insertQuery, values);
        const reclamo = result.rows[0];

        await client.query('COMMIT');

        // Función interna para manejar el envío con un posible retraso
        const procesarEnvio = async () => {
            if (process.env.NODE_ENV === 'test') {
                // Espera 2 segundos antes de enviar en modo test para no saturar el SMTP
                await new Promise(resolve => setTimeout(resolve, 2000));
            }

            return enviarEmails(reclamo, {
                tipo_solicitud,
                nombre_completo,
                email,
                tipo_bien: tipo_bien || 'SERVICIO',
                descripcion_bien,
                detalle_reclamo,
                acepta_copia
            });
        };

        procesarEnvio().catch(err => console.error('Error enviando emails:', err));

        res.status(201).json({
            success: true,
            message: 'Reclamo registrado exitosamente',
            data: {
                codigo_reclamo: reclamo.codigo_reclamo,
                fecha_registro: reclamo.fecha_registro,
                fecha_limite_respuesta: reclamo.fecha_limite_respuesta,
                plazo_dias: 15
            }
        });

    } catch (error) {
        await client.query('ROLLBACK');
        console.error('Error creando reclamo:', error);

        res.status(400).json({
            success: false,
            message: error.message || 'Error al registrar el reclamo',
            error: process.env.NODE_ENV === 'development' ? error.stack : undefined
        });
    } finally {
        client.release();
    }
});

// =============================================================================
// GET /api/reclamos/:codigo - Consultar reclamo por código
// =============================================================================
app.get('/api/reclamos/:codigo', async (req, res) => {
    try {
        const { codigo } = req.params;

        const query = `
            SELECT 
                r.*,
                res.respuesta_empresa,
                res.fecha_respuesta,
                res.respondido_por
            FROM reclamos r
            LEFT JOIN respuestas res ON r.id = res.reclamo_id
            WHERE r.codigo_reclamo = $1
        `;

        const result = await pool.query(query, [codigo]);

        if (result.rows.length === 0) {
            return res.status(404).json({
                success: false,
                message: 'Reclamo no encontrado'
            });
        }

        // Ocultar información sensible
        const reclamo = result.rows[0];
        delete reclamo.firma_digital;
        delete reclamo.ip_address;
        delete reclamo.user_agent;

        res.json({
            success: true,
            data: reclamo
        });

    } catch (error) {
        console.error('Error consultando reclamo:', error);
        res.status(500).json({
            success: false,
            message: 'Error al consultar el reclamo'
        });
    }
});

// =============================================================================
// GET /api/reclamos/:codigo/firma - Ver firma digital
// =============================================================================
app.get('/api/reclamos/:codigo/firma', async (req, res) => {
    try {
        const { codigo } = req.params;

        const result = await pool.query(
            'SELECT firma_digital FROM reclamos WHERE codigo_reclamo = $1',
            [codigo]
        );

        if (result.rows.length === 0) {
            return res.status(404).send('Reclamo no encontrado');
        }

        const firma = result.rows[0].firma_digital;
        const base64Data = firma.replace(/^data:image\/\w+;base64,/, '');
        const buffer = Buffer.from(base64Data, 'base64');

        res.setHeader('Content-Type', 'image/png');
        res.send(buffer);

    } catch (error) {
        res.status(500).send('Error al obtener firma');
    }
});

// =============================================================================
// GET /api/dashboard - Dashboard de reclamos
// =============================================================================
app.get('/api/dashboard', async (req, res) => {
    try {
        const statsQuery = 'SELECT * FROM dashboard_reclamos';
        const pendientesQuery = 'SELECT * FROM reclamos_pendientes LIMIT 10';

        const [stats, pendientes] = await Promise.all([
            pool.query(statsQuery),
            pool.query(pendientesQuery)
        ]);

        res.json({
            success: true,
            data: {
                estadisticas: stats.rows[0],
                pendientes: pendientes.rows
            }
        });

    } catch (error) {
        console.error('Error en dashboard:', error);
        res.status(500).json({
            success: false,
            message: 'Error al obtener dashboard'
        });
    }
});

// =============================================================================
// FUNCIONES DE ENVÍO DE EMAILS
// =============================================================================

async function enviarEmails(reclamo, datos) {
    const { codigo_reclamo, fecha_registro, fecha_limite_respuesta } = reclamo;
    const { tipo_solicitud, nombre_completo, email, tipo_bien, descripcion_bien, detalle_reclamo, acepta_copia } = datos;

    // Email a CODEPLEX (soporte)
    const emailSoporte = {
        from: process.env.SMTP_FROM || 'libro.reclamaciones@codeplex.pe',
        to: process.env.EMAIL_SOPORTE || 'soporte@codeplex.pe',
        subject: `Nuevo ${tipo_solicitud} - ${codigo_reclamo}`,
        html: `
            <!DOCTYPE html>
            <html>
            <head>
                <meta charset="UTF-8">
                <style>
                    body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
                    .container { max-width: 600px; margin: 0 auto; padding: 20px; }
                    .header { background: #1e40af; color: white; padding: 20px; text-align: center; }
                    .content { background: #f9fafb; padding: 20px; border: 1px solid #e5e7eb; }
                    .field { margin-bottom: 15px; }
                    .label { font-weight: bold; color: #1e40af; }
                    .alert { background: #fef2f2; border-left: 4px solid #dc2626; padding: 15px; margin: 20px 0; }
                </style>
            </head>
            <body>
                <div class="container">
                    <div class="header">
                        <h1>📋 Nuevo ${tipo_solicitud}</h1>
                        <p>Código: ${codigo_reclamo}</p>
                    </div>
                    
                    <div class="content">
                        <div class="alert">
                            <strong>⚠️ PLAZO LEGAL:</strong> Debe responder antes del <strong>${new Date(fecha_limite_respuesta).toLocaleDateString('es-PE')}</strong> (15 días hábiles)
                        </div>
                        
                        <div class="field">
                            <div class="label">Tipo de Solicitud:</div>
                            <div>${tipo_solicitud}</div>
                        </div>
                        
                        <div class="field">
                            <div class="label">Consumidor:</div>
                            <div>${nombre_completo}</div>
                        </div>
                        
                        <div class="field">
                            <div class="label">Email:</div>
                            <div>${email}</div>
                        </div>
                        
                       <div class="field" style="background: #fff; padding: 10px; border-radius: 5px; border: 1px solid #e5e7eb;">
                            <div style="margin-bottom: 8px;">
                                <span class="label">Tipo de Bien:</span> 
                                <span style="background-color: ${tipo_bien === 'PRODUCTO' ? '#dbeafe' : '#ffedd5'}; color: ${tipo_bien === 'PRODUCTO' ? '#1e40af' : '#9a3412'}; padding: 2px 6px; border-radius: 4px; font-size: 0.9em; font-weight: bold;">
                                    ${tipo_bien}
                                </span>
                            </div>
                            <div>
                                <span class="label">Descripción:</span><br>
                                ${descripcion_bien}
                            </div>
                            ${monto_reclamado ? `<div style="margin-top: 8px; font-size: 0.9em;"><strong>Monto:</strong> S/ ${monto_reclamado}</div>` : ''}
                            </div>
                        
                       <div class="field">
                            <div class="label">Detalle:</div>
                            <div>${detalle_reclamo}</div>
                        </div>
                        
                      <div class="field">
                            <div class="label">🖊️ Firma Digital:</div>
                            <div>
                                <a href="${process.env.BACKEND_URL || 'http://localhost:3000'}/api/reclamos/${codigo_reclamo}/firma" 
                                   target="_blank" 
                                   style="color: #1e40af; text-decoration: underline;">
                                   Ver firma del consumidor
                                </a>
                            </div>
                        </div>
                    </div>
                </div>
            </body>
            </html>
        `
    };

    await transporter.sendMail(emailSoporte);

    // Email al consumidor (si aceptó recibir copia)
    if (acepta_copia) {
        const emailCliente = {
            from: process.env.SMTP_FROM || 'libro.reclamaciones@codeplex.pe',
            to: email,
            subject: `Confirmación de ${tipo_solicitud} - ${codigo_reclamo}`,
            html: `
                <!DOCTYPE html>
                <html>
                <head>
                    <meta charset="UTF-8">
                    <style>
                        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
                        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
                        .header { background: #1e40af; color: white; padding: 20px; text-align: center; }
                        .success { background: #f0fdf4; border-left: 4px solid #10b981; padding: 15px; margin: 20px 0; }
                        .info { background: #eff6ff; border-left: 4px solid #3b82f6; padding: 15px; margin: 20px 0; }
                    </style>
                </head>
                <body>
                    <div class="container">
                        <div class="header">
                            <h1>✅ ${tipo_solicitud} Registrado</h1>
                            <h2>${codigo_reclamo}</h2>
                        </div>
                        
                        <div class="success">
                            <strong>Estimado/a ${nombre_completo}</strong><br>
                            Su ${tipo_solicitud.toLowerCase()} ha sido registrado exitosamente en nuestro Libro de Reclamaciones Virtual.
                        </div>
                        
                        <div class="info">
                            <strong>📅 Plazo de Respuesta:</strong><br>
                            Recibirá nuestra respuesta antes del <strong>${new Date(fecha_limite_respuesta).toLocaleDateString('es-PE')}</strong>
                            <br><br>
                            Plazo legal: <strong>15 días hábiles</strong> (según D.S. 011-2011-PCM)
                        </div>
                        
                        <h3>Resumen de su ${tipo_solicitud}:</h3>
                        <ul>
                            <li><strong>Código:</strong> ${codigo_reclamo}</li>
                            <li><strong>Fecha:</strong> ${new Date(fecha_registro).toLocaleString('es-PE')}</li>
                            <li><strong>Tipo:</strong> ${tipo_solicitud}</li>
                           <li><strong>Tipo de Solicitud:</strong> ${tipo_solicitud}</li>
                            <li><strong>Bien Contratado:</strong> ${tipo_bien}</li>
                            <li><strong>Descripción:</strong> ${descripcion_bien}</li>
                        </ul>
                        
                        <div class="info">
                            <strong>ℹ️ Información Importante:</strong><br>
                            • Conserve este email como comprobante<br>
                        </div>
                        
                        <div style="text-align: center; margin-top: 30px; padding-top: 20px; border-top: 1px solid #e5e7eb; color: #6b7280; font-size: 12px;">
                            CODEPLEX SAC | RUC: 20539782232<br>
                            AV. LOS PROCERES MZA. G3 LOTE. 11 - LIMA - LOS OLIVOS<br>
                            soporte@codeplex.pe | +51 936343607
                        </div>
                    </div>
                </body>
                </html>
            `
        };

        await transporter.sendMail(emailCliente);
    }
}

// =============================================================================
// MANEJO DE ERRORES GLOBAL
// =============================================================================
app.use((err, req, res, next) => {
    console.error('Error no manejado:', err);
    res.status(500).json({
        success: false,
        message: 'Error interno del servidor',
        error: process.env.NODE_ENV === 'development' ? err.message : undefined
    });
});

// =============================================================================
// INICIAR SERVIDOR
// =============================================================================
app.listen(PORT, () => {
    console.log(`🚀 Servidor corriendo en puerto ${PORT}`);
    console.log(`📍 Health check: http://localhost:${PORT}/api/health`);
});

// Manejo de cierre graceful
process.on('SIGTERM', async () => {
    console.log('👋 Cerrando servidor...');
    await pool.end();
    process.exit(0);
});
