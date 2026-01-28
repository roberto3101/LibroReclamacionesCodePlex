-- =============================================================================
-- SCHEMA UNIVERSAL - LIBRO DE RECLAMACIONES CODEPLEX
-- Compatible con PostgreSQL Y CockroachDB
-- Conforme a D.S. 011-2011-PCM y Ley N° 29571
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
    codigo_reclamo VARCHAR(50) UNIQUE NOT NULL,
    tipo_solicitud VARCHAR(20) NOT NULL CHECK (tipo_solicitud IN ('RECLAMO', 'QUEJA')),
    estado VARCHAR(20) DEFAULT 'PENDIENTE' CHECK (estado IN ('PENDIENTE', 'EN_PROCESO', 'RESUELTO', 'CERRADO')),
    
    -- Datos del consumidor (Sección 1 - Obligatorio D.S. 011-2011-PCM)
    nombre_completo VARCHAR(255) NOT NULL,
    tipo_documento VARCHAR(20) NOT NULL CHECK (tipo_documento IN ('DNI', 'CE', 'Pasaporte', 'RUC')),
    numero_documento VARCHAR(50) NOT NULL,
    telefono VARCHAR(50) NOT NULL,
    email VARCHAR(255) NOT NULL,
    domicilio TEXT,
    departamento VARCHAR(100),
    provincia VARCHAR(100),
    distrito VARCHAR(100),
    
    -- Datos del proveedor (Sección 2 - Auto-completado)
    razon_social VARCHAR(255) DEFAULT 'CODEPLEX SOFTWARE S.A.C.',
    ruc VARCHAR(11) DEFAULT '20539782232',
    direccion_proveedor TEXT DEFAULT 'Jr. Las Colcas Mza. R Lote 8 - Urb. Los Naranjos - Los Olivos, Lima',
    
    -- Bien contratado o producto (Sección 3)
    tipo_bien VARCHAR(20) CHECK (tipo_bien IN ('PRODUCTO', 'SERVICIO')),
    monto_reclamado DECIMAL(10, 2) DEFAULT 0,
    descripcion_bien TEXT NOT NULL,
    
    -- Campos específicos para QUEJA
    area_queja VARCHAR(255),
    descripcion_situacion TEXT,
    
    -- Detalle del reclamo/queja (Sección 4)
    fecha_incidente DATE NOT NULL,
    detalle_reclamo TEXT NOT NULL,
    pedido_consumidor TEXT NOT NULL,
    
    -- Firma digital (Canvas base64)
    firma_digital TEXT NOT NULL,
    ip_address INET,
    user_agent TEXT,
    
    -- Conformidad legal
    acepta_terminos BOOLEAN DEFAULT true,
    acepta_copia BOOLEAN DEFAULT true,
    
    -- Fechas de control
    fecha_registro TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    fecha_limite_respuesta DATE GENERATED ALWAYS AS (fecha_registro::DATE + INTERVAL '15 days') STORED,
    fecha_respuesta TIMESTAMP
);

-- Índices para optimizar búsquedas
CREATE INDEX idx_codigo ON reclamos(codigo_reclamo);
CREATE INDEX idx_estado ON reclamos(estado);
CREATE INDEX idx_tipo ON reclamos(tipo_solicitud);
CREATE INDEX idx_fecha ON reclamos(fecha_registro DESC);
CREATE INDEX idx_email ON reclamos(email);
CREATE INDEX idx_documento ON reclamos(numero_documento);

-- =============================================================================
-- TABLA: respuestas
-- Almacena las respuestas de CODEPLEX a cada reclamo
-- =============================================================================
CREATE TABLE respuestas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reclamo_id UUID NOT NULL REFERENCES reclamos(id) ON DELETE CASCADE,
    
    -- Contenido de la respuesta
    respuesta_empresa TEXT NOT NULL,
    accion_tomada TEXT,
    compensacion_ofrecida TEXT,
    
    -- Usuario que respondió
    respondido_por VARCHAR(255) NOT NULL,
    cargo_responsable VARCHAR(100),
    
    -- Control de fechas
    fecha_respuesta TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Archivos adjuntos (opcional - JSON)
    archivos_adjuntos JSONB,
    
    -- Notificación al cliente
    notificado_cliente BOOLEAN DEFAULT false,
    fecha_notificacion TIMESTAMP
);

-- Índices
CREATE INDEX idx_reclamo ON respuestas(reclamo_id);
CREATE INDEX idx_fecha_respuesta ON respuestas(fecha_respuesta DESC);

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
    (fecha_limite_respuesta - CURRENT_DATE) AS dias_restantes,
    CASE 
        WHEN fecha_limite_respuesta < CURRENT_DATE THEN 'VENCIDO'
        WHEN fecha_limite_respuesta - CURRENT_DATE <= 3 THEN 'URGENTE'
        ELSE 'NORMAL'
    END AS prioridad
FROM reclamos
WHERE estado = 'PENDIENTE'
ORDER BY fecha_limite_respuesta ASC;

-- =============================================================================
-- FIN DEL SCHEMA - Compatible con PostgreSQL Y CockroachDB
-- =============================================================================