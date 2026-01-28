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








-- =============================================================================
-- TABLAS ADICIONALES PARA SEGUIMIENTO Y TRAZABILIDAD
-- Compatible con PostgreSQL Y CockroachDB
-- =============================================================================

-- Tabla de historial/trazabilidad de cambios de estado
CREATE TABLE historial_reclamos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reclamo_id UUID NOT NULL REFERENCES reclamos(id) ON DELETE CASCADE,
    
    estado_anterior VARCHAR(20),
    estado_nuevo VARCHAR(20) NOT NULL,
    
    comentario TEXT,
    usuario_accion VARCHAR(255) DEFAULT 'SISTEMA',
    tipo_accion VARCHAR(50) NOT NULL CHECK (tipo_accion IN ('CREACION', 'CAMBIO_ESTADO', 'RESPUESTA', 'MENSAJE_CLIENTE', 'ADJUNTO', 'NOTIFICACION')),
    
    ip_address INET,
    user_agent TEXT,
    
    fecha_accion TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_historial_reclamo ON historial_reclamos(reclamo_id);
CREATE INDEX idx_historial_fecha ON historial_reclamos(fecha_accion DESC);

-- Tabla de mensajes adicionales del cliente (seguimiento)
CREATE TABLE mensajes_seguimiento (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reclamo_id UUID NOT NULL REFERENCES reclamos(id) ON DELETE CASCADE,
    
    tipo_mensaje VARCHAR(20) NOT NULL CHECK (tipo_mensaje IN ('CLIENTE', 'EMPRESA')),
    mensaje TEXT NOT NULL,
    
    archivo_adjunto TEXT,
    nombre_archivo VARCHAR(255),
    
    leido BOOLEAN DEFAULT false,
    fecha_lectura TIMESTAMP,
    
    fecha_mensaje TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_mensajes_reclamo ON mensajes_seguimiento(reclamo_id);
CREATE INDEX idx_mensajes_fecha ON mensajes_seguimiento(fecha_mensaje DESC);

-- Vista de seguimiento completo
CREATE VIEW seguimiento_completo AS
SELECT 
    r.id,
    r.codigo_reclamo,
    r.tipo_solicitud,
    r.estado,
    r.nombre_completo,
    r.numero_documento,
    r.email,
    r.telefono,
    r.descripcion_bien,
    r.detalle_reclamo,
    r.pedido_consumidor,
    r.fecha_registro,
    r.fecha_limite_respuesta,
    r.fecha_respuesta,
    (r.fecha_limite_respuesta - CURRENT_DATE) AS dias_restantes,
    CASE 
        WHEN r.estado = 'RESUELTO' OR r.estado = 'CERRADO' THEN 'COMPLETADO'
        WHEN r.fecha_limite_respuesta < CURRENT_DATE THEN 'VENCIDO'
        WHEN r.fecha_limite_respuesta - CURRENT_DATE <= 3 THEN 'URGENTE'
        ELSE 'EN_TIEMPO'
    END AS prioridad,
    res.respuesta_empresa,
    res.accion_tomada,
    res.compensacion_ofrecida,
    res.respondido_por,
    res.fecha_respuesta AS fecha_respuesta_empresa
FROM reclamos r
LEFT JOIN respuestas res ON r.id = res.reclamo_id;

-- Trigger para registrar historial automáticamente (PostgreSQL)
-- Nota: CockroachDB no soporta triggers, pero el backend lo manejará
CREATE OR REPLACE FUNCTION registrar_historial_estado()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.estado IS DISTINCT FROM NEW.estado THEN
        INSERT INTO historial_reclamos (reclamo_id, estado_anterior, estado_nuevo, tipo_accion, comentario)
        VALUES (NEW.id, OLD.estado, NEW.estado, 'CAMBIO_ESTADO', 'Cambio automático de estado');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Solo ejecutar si es PostgreSQL (CockroachDB ignorará esto)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'registrar_historial_estado') THEN
        DROP TRIGGER IF EXISTS trigger_historial_estado ON reclamos;
        CREATE TRIGGER trigger_historial_estado
            AFTER UPDATE ON reclamos
            FOR EACH ROW
            EXECUTE FUNCTION registrar_historial_estado();
    END IF;
EXCEPTION WHEN OTHERS THEN
    NULL;
END $$;











