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
    fecha_respuesta TIMESTAMP,

    -- Gestión Interna (Relación con tabla usuarios_admin)
    atendido_por UUID REFERENCES usuarios_admin(id) ON DELETE SET NULL
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









































-- ============================================================================
-- SCHEMA EXTENSION: Sistema de Usuarios Administradores
-- Para: Panel de Gestión de Reclamos
-- Base de datos: PostgreSQL / CockroachDB
-- Fecha: 2026-01-27
-- ============================================================================

-- ============================================================================
-- TABLA: usuarios_admin
-- Descripción: Usuarios que pueden gestionar los reclamos desde el panel admin
-- Roles: ADMIN (full access), SOPORTE (limited access)
-- ============================================================================
CREATE TABLE IF NOT EXISTS usuarios_admin (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    nombre_completo VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    rol VARCHAR(50) NOT NULL CHECK (rol IN ('ADMIN', 'SOPORTE')),
    activo BOOLEAN DEFAULT true NOT NULL,
    debe_cambiar_password BOOLEAN DEFAULT true NOT NULL,
    ultimo_acceso TIMESTAMP,
    fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    creado_por UUID REFERENCES usuarios_admin(id) ON DELETE SET NULL,
    
    -- Constraints
    CONSTRAINT email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

-- ============================================================================
-- ÍNDICES para optimizar consultas
-- ============================================================================
CREATE INDEX IF NOT EXISTS idx_usuarios_admin_email ON usuarios_admin(email);
CREATE INDEX IF NOT EXISTS idx_usuarios_admin_activo ON usuarios_admin(activo);
CREATE INDEX IF NOT EXISTS idx_usuarios_admin_rol ON usuarios_admin(rol);
CREATE INDEX IF NOT EXISTS idx_usuarios_admin_creacion ON usuarios_admin(fecha_creacion DESC);

-- ============================================================================
-- TABLA: sesiones_admin (opcional - para invalidar tokens)
-- Descripción: Registro de sesiones activas para poder invalidar tokens JWT
-- ============================================================================
CREATE TABLE IF NOT EXISTS sesiones_admin (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID NOT NULL REFERENCES usuarios_admin(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    fecha_inicio TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    fecha_expiracion TIMESTAMP NOT NULL,
    activa BOOLEAN DEFAULT true NOT NULL,
    
    -- Constraints
    CONSTRAINT sesion_valida CHECK (fecha_expiracion > fecha_inicio)
);

CREATE INDEX IF NOT EXISTS idx_sesiones_usuario ON sesiones_admin(usuario_id);
CREATE INDEX IF NOT EXISTS idx_sesiones_token ON sesiones_admin(token_hash);
CREATE INDEX IF NOT EXISTS idx_sesiones_activas ON sesiones_admin(activa, fecha_expiracion);

-- ============================================================================
-- TABLA: auditoria_admin
-- Descripción: Log de acciones realizadas por administradores
-- ============================================================================
CREATE TABLE IF NOT EXISTS auditoria_admin (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID NOT NULL REFERENCES usuarios_admin(id) ON DELETE CASCADE,
    accion VARCHAR(100) NOT NULL,
    entidad VARCHAR(50) NOT NULL,
    entidad_id VARCHAR(255),
    detalles JSONB,
    ip_address VARCHAR(45),
    fecha TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_auditoria_usuario ON auditoria_admin(usuario_id);
CREATE INDEX IF NOT EXISTS idx_auditoria_fecha ON auditoria_admin(fecha DESC);
CREATE INDEX IF NOT EXISTS idx_auditoria_accion ON auditoria_admin(accion);
CREATE INDEX IF NOT EXISTS idx_auditoria_entidad ON auditoria_admin(entidad, entidad_id);

-- ============================================================================
-- VISTA: vista_usuarios_admin
-- Descripción: Vista con información completa de usuarios incluyendo creador
-- ============================================================================
CREATE OR REPLACE VIEW vista_usuarios_admin AS
SELECT 
    u.id,
    u.email,
    u.nombre_completo,
    u.rol,
    u.activo,
    u.debe_cambiar_password,
    u.fecha_creacion,
    u.ultimo_acceso,
    creador.nombre_completo as creado_por_nombre,
    creador.email as creado_por_email,
    -- Estadísticas
    COALESCE(sesiones_count.total, 0) as total_sesiones,
    COALESCE(acciones_count.total, 0) as total_acciones
FROM usuarios_admin u
LEFT JOIN usuarios_admin creador ON u.creado_por = creador.id
LEFT JOIN (
    SELECT usuario_id, COUNT(*) as total
    FROM sesiones_admin
    GROUP BY usuario_id
) sesiones_count ON u.id = sesiones_count.usuario_id
LEFT JOIN (
    SELECT usuario_id, COUNT(*) as total
    FROM auditoria_admin
    GROUP BY usuario_id
) acciones_count ON u.id = acciones_count.usuario_id;

-- ============================================================================
-- VISTA: estadisticas_simples
-- Descripción: Estadísticas básicas del sistema de reclamos
-- ============================================================================
CREATE OR REPLACE VIEW estadisticas_simples AS
SELECT 
    -- Total de reclamos
    COUNT(*) as total_reclamos,
    
    -- Por estado
    COUNT(*) FILTER (WHERE estado = 'PENDIENTE') as pendientes,
    COUNT(*) FILTER (WHERE estado = 'EN_PROCESO') as en_proceso,
    COUNT(*) FILTER (WHERE estado = 'RESUELTO') as resueltos,
    COUNT(*) FILTER (WHERE estado = 'CERRADO') as cerrados,
    
    -- Hoy
    COUNT(*) FILTER (WHERE DATE(fecha_registro) = CURRENT_DATE) as reclamos_hoy,
    
    -- Esta semana
    COUNT(*) FILTER (WHERE fecha_registro >= CURRENT_DATE - INTERVAL '7 days') as reclamos_semana,
    
    -- Este mes
    COUNT(*) FILTER (WHERE fecha_registro >= DATE_TRUNC('month', CURRENT_DATE)) as reclamos_mes,
    
    -- Promedio de días para resolver
    ROUND(AVG(
        CASE 
            WHEN estado IN ('RESUELTO', 'CERRADO') AND fecha_respuesta IS NOT NULL
            THEN EXTRACT(EPOCH FROM (fecha_respuesta - fecha_registro)) / 86400
            ELSE NULL
        END
    ), 1) as promedio_dias_resolucion
FROM reclamos;

-- ============================================================================
-- FUNCIÓN: registrar_auditoria
-- Descripción: Función helper para registrar acciones en auditoría
-- ============================================================================
CREATE OR REPLACE FUNCTION registrar_auditoria(
    p_usuario_id UUID,
    p_accion VARCHAR,
    p_entidad VARCHAR,
    p_entidad_id VARCHAR,
    p_detalles JSONB DEFAULT NULL,
    p_ip_address VARCHAR DEFAULT NULL
)
RETURNS UUID
LANGUAGE plpgsql
AS $$
DECLARE
    v_auditoria_id UUID;
BEGIN
    INSERT INTO auditoria_admin (
        usuario_id,
        accion,
        entidad,
        entidad_id,
        detalles,
        ip_address
    ) VALUES (
        p_usuario_id,
        p_accion,
        p_entidad,
        p_entidad_id,
        p_detalles,
        p_ip_address
    )
    RETURNING id INTO v_auditoria_id;
    
    RETURN v_auditoria_id;
END;
$$;

-- ============================================================================
-- COMENTARIOS en las tablas
-- ============================================================================
COMMENT ON TABLE usuarios_admin IS 'Usuarios administradores del sistema de reclamos';
COMMENT ON TABLE sesiones_admin IS 'Registro de sesiones activas para control de tokens JWT';
COMMENT ON TABLE auditoria_admin IS 'Log de auditoría de todas las acciones administrativas';

COMMENT ON COLUMN usuarios_admin.debe_cambiar_password IS 'Flag para forzar cambio de contraseña en el siguiente login';
COMMENT ON COLUMN usuarios_admin.creado_por IS 'Usuario administrador que creó esta cuenta';
COMMENT ON COLUMN sesiones_admin.token_hash IS 'Hash SHA256 del token JWT para invalidación';
COMMENT ON COLUMN auditoria_admin.detalles IS 'Información adicional en formato JSON sobre la acción realizada';

-- ============================================================================
-- INSERTAR PRIMER USUARIO ADMIN
-- ============================================================================
-- IMPORTANTE: Este usuario inicial NO tiene creado_por porque es el primero
-- Password temporal: "Admin123!" (debes cambiar este hash con bcrypt en tu backend)
-- Este es solo un placeholder - el backend generará el hash real

-- NOTA: Ejecuta esto MANUALMENTE después de tener tu backend funcionando
-- para generar el hash correcto con bcrypt

/*
INSERT INTO usuarios_admin (
    email,
    nombre_completo,
    password_hash,
    rol,
    debe_cambiar_password,
    activo,
    creado_por
) VALUES (
    'admin@codeplex.com',
    'Administrador Principal',
    '$2a$10$PLACEHOLDER_HASH_CAMBIAR_ESTO',
    'ADMIN',
    false,  -- No forzar cambio para el primer admin
    true,
    NULL    -- Primer usuario no tiene creador
);
*/

-- ============================================================================
-- GRANTS (ajusta según tus necesidades de seguridad)
-- ============================================================================
-- Si tienes un usuario específico para la aplicación:
-- GRANT SELECT, INSERT, UPDATE ON usuarios_admin TO app_user;
-- GRANT SELECT, INSERT, UPDATE ON sesiones_admin TO app_user;
-- GRANT SELECT, INSERT ON auditoria_admin TO app_user;
-- GRANT SELECT ON vista_usuarios_admin TO app_user;
-- GRANT SELECT ON estadisticas_simples TO app_user;

-- ============================================================================
-- FIN DEL SCHEMA
-- ============================================================================

-- Para verificar que todo se creó correctamente:
-- SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name LIKE '%admin%';

