-- =============================================================================
-- SCHEMA COCKROACHDB - LIBRO DE RECLAMACIONES CODEPLEX
-- Conforme a D.S. 011-2011-PCM y Ley N° 29571
-- Compatible con CockroachDB Serverless
-- =============================================================================

-- Eliminar tablas si existen
DROP TABLE IF EXISTS respuestas CASCADE;
DROP TABLE IF EXISTS reclamos CASCADE;

-- =============================================================================
-- TABLA PRINCIPAL: reclamos
-- Almacena todos los reclamos y quejas según normativa INDECOPI
-- =============================================================================
CREATE TABLE reclamos (
    -- Identificación del reclamo (UUID para distributed DB)
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    codigo_reclamo STRING UNIQUE NOT NULL,
    tipo_solicitud STRING NOT NULL CHECK (tipo_solicitud IN ('RECLAMO', 'QUEJA')),
    estado STRING DEFAULT 'PENDIENTE' CHECK (estado IN ('PENDIENTE', 'EN_PROCESO', 'RESUELTO', 'CERRADO')),
    
    -- Datos del consumidor (Sección 1 - Obligatorio D.S. 011-2011-PCM)
    nombre_completo STRING NOT NULL,
    tipo_documento STRING NOT NULL CHECK (tipo_documento IN ('DNI', 'CE', 'Pasaporte', 'RUC')),
    numero_documento STRING NOT NULL,
    telefono STRING NOT NULL,
    email STRING NOT NULL,
    domicilio STRING,
    departamento STRING,
    provincia STRING,
    distrito STRING,
    
    -- Datos del proveedor (Sección 2 - Auto-completado)
    razon_social STRING DEFAULT 'CODEPLEX SAC',
    ruc STRING DEFAULT '20XXXXXXXXX',
    direccion_proveedor STRING DEFAULT 'AV. LOS PROCERES MZA. G3 LOTE. 11 - LIMA - LOS OLIVOS',
    
    -- Bien contratado o producto (Sección 3)
    tipo_bien STRING CHECK (tipo_bien IN ('PRODUCTO', 'SERVICIO')),
    monto_reclamado DECIMAL(10, 2) DEFAULT 0,
    descripcion_bien STRING NOT NULL,
    
    -- Campos específicos para QUEJA
    area_queja STRING,
    descripcion_situacion STRING,
    
    -- Detalle del reclamo/queja (Sección 4)
    fecha_incidente DATE NOT NULL,
    detalle_reclamo STRING NOT NULL,
    pedido_consumidor STRING NOT NULL,
    
    -- Firma digital (Canvas base64)
    firma_digital STRING NOT NULL,
    ip_address INET,
    user_agent STRING,
    
    -- Conformidad legal
    acepta_terminos BOOL DEFAULT true,
    acepta_copia BOOL DEFAULT true,
    
    -- Fechas de control
    fecha_registro TIMESTAMP DEFAULT current_timestamp(),
    fecha_limite_respuesta DATE AS (fecha_registro::DATE + INTERVAL '15 days') STORED,
    fecha_respuesta TIMESTAMP,
    
    -- Índices para optimizar búsquedas
    INDEX idx_codigo (codigo_reclamo),
    INDEX idx_estado (estado),
    INDEX idx_tipo (tipo_solicitud),
    INDEX idx_fecha (fecha_registro DESC),
    INDEX idx_email (email),
    INDEX idx_documento (numero_documento)
);

-- =============================================================================
-- TABLA: respuestas
-- Almacena las respuestas de CODEPLEX a cada reclamo
-- =============================================================================
CREATE TABLE respuestas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reclamo_id UUID NOT NULL REFERENCES reclamos(id) ON DELETE CASCADE,
    
    -- Contenido de la respuesta
    respuesta_empresa STRING NOT NULL,
    accion_tomada STRING,
    compensacion_ofrecida STRING,
    
    -- Usuario que respondió
    respondido_por STRING NOT NULL,
    cargo_responsable STRING,
    
    -- Control de fechas
    fecha_respuesta TIMESTAMP DEFAULT current_timestamp(),
    
    -- Archivos adjuntos (opcional - JSON)
    archivos_adjuntos JSONB,
    
    -- Notificación al cliente
    notificado_cliente BOOL DEFAULT false,
    fecha_notificacion TIMESTAMP,
    
    INDEX idx_reclamo (reclamo_id),
    INDEX idx_fecha_respuesta (fecha_respuesta DESC)
);

-- =============================================================================
-- VISTA: Dashboard de reclamos
-- =============================================================================
CREATE VIEW dashboard_reclamos AS
SELECT 
    COUNT(*) FILTER (WHERE estado = 'PENDIENTE') AS pendientes,
    COUNT(*) FILTER (WHERE estado = 'EN_PROCESO') AS en_proceso,
    COUNT(*) FILTER (WHERE estado = 'RESUELTO') AS resueltos,
    COUNT(*) FILTER (WHERE fecha_limite_respuesta < CURRENT_DATE AND estado = 'PENDIENTE') AS vencidos,
    COUNT(*) FILTER (WHERE tipo_solicitud = 'RECLAMO') AS total_reclamos,
    COUNT(*) FILTER (WHERE tipo_solicitud = 'QUEJA') AS total_quejas,
    COUNT(*) AS total
FROM reclamos;

-- =============================================================================
-- VISTA: Reclamos pendientes con días restantes
-- =============================================================================
CREATE VIEW reclamos_pendientes AS
SELECT 
    id,
    codigo_reclamo,
    tipo_solicitud,
    nombre_completo,
    email,
    fecha_registro,
    fecha_limite_respuesta,
    (fecha_limite_respuesta - CURRENT_DATE)::INT AS dias_restantes,
    CASE 
        WHEN fecha_limite_respuesta < CURRENT_DATE THEN 'VENCIDO'
        WHEN fecha_limite_respuesta - CURRENT_DATE <= 3 THEN 'URGENTE'
        ELSE 'NORMAL'
    END AS prioridad
FROM reclamos
WHERE estado = 'PENDIENTE'
ORDER BY fecha_limite_respuesta ASC;

-- =============================================================================
-- FIN DEL SCHEMA
-- =============================================================================
