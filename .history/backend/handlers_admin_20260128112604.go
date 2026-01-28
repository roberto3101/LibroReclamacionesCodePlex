package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"     // <--- NUEVO
	"net/http"
	"strconv"  // <--- NUEVO
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Variable global para JWT

// =============================================================================
// HANDLERS DE AUTENTICACIÓN ADMIN
// =============================================================================

func loginAdminHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	// DEBUG 1: Ver si llegan los datos JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("⚠️ Error recibiendo JSON login: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Datos inválidos"})
		return
	}

	log.Printf("🔑 Intento de login para: '%s' (Password len: %d)", req.Email, len(req.Password))

	var id, passwordHash, rol, nombreCompleto string
	var activo, debeCambiarPassword bool

	// DEBUG 2: Buscar usuario ignorando mayúsculas y espacios
	// Usamos LOWER() y TrimSpace() para asegurar que encuentre 'admin@codeplex.com'
	err := pool.QueryRow(ctx, `
		SELECT id, password_hash, rol, activo, nombre_completo, debe_cambiar_password
		FROM usuarios_admin WHERE email = LOWER($1)
	`, strings.TrimSpace(req.Email)).Scan(&id, &passwordHash, &rol, &activo, &nombreCompleto, &debeCambiarPassword)

	if err != nil {
		log.Printf("❌ Usuario NO encontrado en BD: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Credenciales inválidas"})
		return
	}

	if !activo {
		log.Printf("⛔ Usuario inactivo: %s", req.Email)
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "Usuario inactivo"})
		return
	}

	// DEBUG 3: Comparar contraseñas
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		log.Printf("❌ Contraseña INCORRECTA para %s. Error: %v", req.Email, err)
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Credenciales inválidas"})
		return
	}

	log.Printf("✅ Login EXITOSO para: %s", req.Email)

	// Actualizar último acceso
	go pool.Exec(context.Background(), "UPDATE usuarios_admin SET ultimo_acceso = NOW() WHERE id = $1", id)

	// Generar token JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": id,
		"email":   req.Email,
		"rol":     rol,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Printf("⚠️ Error firmando token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error interno"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"token": tokenString,
			"usuario": gin.H{
				"id":                    id,
				"email":                 req.Email,
				"nombre_completo":       nombreCompleto,
				"rol":                   rol,
				"debe_cambiar_password": debeCambiarPassword,
			},
			"expires_in": 86400,
		},
	})
}

// =============================================================================
// MIDDLEWARE JWT
// =============================================================================

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Token requerido"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Token inválido"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_id", claims["user_id"])
			c.Set("email", claims["email"])
			c.Set("rol", claims["rol"])
		}
		c.Next()
	}
}

func rolAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rol, _ := c.Get("rol")
		if rol != "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "Acceso denegado"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// =============================================================================
// HANDLERS DE RECLAMOS ADMIN
// =============================================================================

func listarReclamosAdminHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	// 1. Obtener parámetros de paginación
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30")) // 30 registros por defecto
	estado := c.Query("estado")
	search := c.Query("search") // Opcional: Preparado para búsqueda futura

	if page < 1 { page = 1 }
	if limit < 1 { limit = 30 }
	offset := (page - 1) * limit

	// 2. Construir filtros dinámicos (WHERE)
	whereClause := " WHERE 1=1"
	args := []interface{}{}
	argID := 1

	if estado != "" {
		whereClause += fmt.Sprintf(" AND r.estado = $%d", argID)
		args = append(args, estado)
		argID++
	}

	if search != "" {
		searchPattern := "%" + search + "%"
		whereClause += fmt.Sprintf(" AND (r.codigo_reclamo ILIKE $%d OR r.nombre_completo ILIKE $%d OR r.email ILIKE $%d)", argID, argID, argID)
		args = append(args, searchPattern) // Se usa el mismo argumento 3 veces en el OR, pero PG prefiere índices únicos si es driver estricto, aquí simplificamos uso secuencial
		argID++
	}

	// 3. CONSULTA 1: Contar el total de registros (para saber cuántas páginas hay)
	var total int64
	countQuery := "SELECT COUNT(*) FROM reclamos r" + whereClause
	
	// Nota: Si usas el driver pgx puro a veces requiere reajuste de argumentos posicionales ($1, $2...) 
	// Para simplificar, asumimos que el filtro de búsqueda simple no choca.
	// Si search se complica, mejor hacerlo condicional.
	
	err := pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error contando registros"})
		return
	}

	// 4. CONSULTA 2: Obtener los datos paginados
	query := `
		SELECT r.id, r.codigo_reclamo, r.tipo_solicitud, r.estado, r.nombre_completo, 
			   r.email, r.telefono, r.descripcion_bien, r.fecha_registro, 
			   r.fecha_limite_respuesta,
			   (r.fecha_limite_respuesta - CURRENT_DATE)::int AS dias_restantes,
			   ua.nombre_completo as nombre_admin_atendio
		FROM reclamos r
		LEFT JOIN usuarios_admin ua ON r.atendido_por = ua.id
	` + whereClause + fmt.Sprintf(" ORDER BY r.fecha_registro DESC LIMIT $%d OFFSET $%d", argID, argID+1)
	
	// Agregamos limit y offset a los argumentos
	args = append(args, limit, offset)

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error al consultar reclamos"})
		return
	}
	defer rows.Close()

	var reclamos []gin.H
	for rows.Next() {
		var r struct {
			ID, CodigoReclamo, TipoSolicitud, Estado, NombreCompleto, Email, Telefono, DescripcionBien string
			FechaRegistro, FechaLimiteRespuesta time.Time
			DiasRestantes *int
			NombreAdminAtendio sql.NullString
		}
		if err := rows.Scan(&r.ID, &r.CodigoReclamo, &r.TipoSolicitud, &r.Estado, &r.NombreCompleto,
			&r.Email, &r.Telefono, &r.DescripcionBien, &r.FechaRegistro, &r.FechaLimiteRespuesta, &r.DiasRestantes, &r.NombreAdminAtendio); err == nil {
			
			reclamos = append(reclamos, gin.H{
				"id":                     r.ID,
				"codigo_reclamo":         r.CodigoReclamo,
				"tipo_solicitud":         r.TipoSolicitud,
				"estado":                 r.Estado,
				"nombre_completo":        r.NombreCompleto,
				"email":                  r.Email,
				"telefono":               r.Telefono,
				"descripcion_bien":       r.DescripcionBien,
				"fecha_registro":         r.FechaRegistro,
				"fecha_limite_respuesta": r.FechaLimiteRespuesta,
				"dias_restantes":         r.DiasRestantes,
				"nombre_admin_atendio":   nullToInterface(r.NombreAdminAtendio),
			})
		}
	}

	if reclamos == nil {
		reclamos = []gin.H{}
	}

	// 5. Calcular metadatos de paginación
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    reclamos,
		"pagination": gin.H{
			"total":        total,
			"page":         page,
			"limit":        limit,
			"total_pages":  totalPages,
			"has_next":     page < totalPages,
			"has_prev":     page > 1,
		},
	})
}

func cambiarEstadoReclamoHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id := c.Param("id")
	userID, _ := c.Get("user_id")
	rol, _ := c.Get("rol")

	var req struct {
		Estado     string `json:"estado" binding:"required,oneof=PENDIENTE EN_PROCESO RESUELTO CERRADO"`
		Comentario string `json:"comentario"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Datos inválidos"})
		return
	}

	// SOPORTE no puede cerrar reclamos
	if rol == "SOPORTE" && req.Estado == "CERRADO" {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "No tiene permisos para cerrar reclamos"})
		return
	}

	var estadoAnterior string
	err := pool.QueryRow(ctx, "SELECT estado FROM reclamos WHERE id = $1", id).Scan(&estadoAnterior)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Reclamo no encontrado"})
		return
	}

	_, err = pool.Exec(ctx, "UPDATE reclamos SET estado = $1 WHERE id = $2", req.Estado, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error actualizando estado"})
		return
	}

	// Registrar en historial
	pool.Exec(ctx, `
		INSERT INTO historial_reclamos (reclamo_id, estado_anterior, estado_nuevo, tipo_accion, comentario, usuario_accion)
		VALUES ($1, $2, $3, 'CAMBIO_ESTADO', $4, $5)
	`, id, estadoAnterior, req.Estado, req.Comentario, userID)

	// Registrar auditoría
	userIDStr := fmt.Sprintf("%v", userID)
	_, errAudit := pool.Exec(ctx, `
		INSERT INTO auditoria_admin (usuario_id, accion, entidad, entidad_id, detalles, ip_address)
		VALUES ($1::uuid, 'CAMBIO_ESTADO', 'RECLAMO', $2, $3::jsonb, $4)
	`, userIDStr, id, fmt.Sprintf(`{"estado_anterior":"%s","estado_nuevo":"%s"}`, estadoAnterior, req.Estado), c.ClientIP())
	if errAudit != nil {
		fmt.Printf("⚠️ Error registrando auditoría: %v\n", errAudit)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Estado actualizado correctamente",
	})
}

func responderReclamoHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id := c.Param("id")
	userID, _ := c.Get("user_id")
	email, _ := c.Get("email")

	var req struct {
		RespuestaEmpresa     string `json:"respuesta_empresa" binding:"required,min=10"`
		AccionTomada         string `json:"accion_tomada"`
		CompensacionOfrecida string `json:"compensacion_ofrecida"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Datos inválidos"})
		return
	}

	_, err := pool.Exec(ctx, `
		INSERT INTO respuestas (reclamo_id, respuesta_empresa, accion_tomada, compensacion_ofrecida, respondido_por)
		VALUES ($1, $2, $3, $4, $5)
	`, id, req.RespuestaEmpresa, nullString(&req.AccionTomada), nullString(&req.CompensacionOfrecida), email)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error guardando respuesta"})
		return
	}

	// Actualizar estado, fecha de respuesta y REGISTRAR QUIÉN ATENDIÓ
    pool.Exec(ctx, "UPDATE reclamos SET estado = 'RESUELTO', fecha_respuesta = NOW(), atendido_por = $2 WHERE id = $1", id, userID)

    // Historial

	// Historial
	pool.Exec(ctx, `
		INSERT INTO historial_reclamos (reclamo_id, estado_anterior, estado_nuevo, tipo_accion, comentario, usuario_accion)
		SELECT estado, 'RESUELTO', 'RESPUESTA', 'Respuesta enviada por la empresa', $2
		FROM reclamos WHERE id = $1
	`, id, email)

	// Auditoría
	userIDStr := fmt.Sprintf("%v", userID)
	_, errAudit := pool.Exec(ctx, `
		INSERT INTO auditoria_admin (usuario_id, accion, entidad, entidad_id, ip_address)
		VALUES ($1::uuid, 'RESPONDER', 'RECLAMO', $2, $3)
	`, userIDStr, id, c.ClientIP())
	if errAudit != nil {
		fmt.Printf("⚠️ Error registrando auditoría: %v\n", errAudit)
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Respuesta enviada correctamente"})
}

func obtenerEstadisticasHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Usamos int64 para coincidir con el BIGINT de PostgreSQL
	var stats struct {
		TotalReclamos, Pendientes, EnProceso, Resueltos, Cerrados int64
		ReclamosHoy, ReclamosSemana, ReclamosMes                  int64
		PromedioDiasResolucion                                    *float64
	}

	err := pool.QueryRow(ctx, "SELECT * FROM estadisticas_simples").Scan(
		&stats.TotalReclamos, &stats.Pendientes, &stats.EnProceso, &stats.Resueltos,
		&stats.Cerrados, &stats.ReclamosHoy, &stats.ReclamosSemana, &stats.ReclamosMes,
		&stats.PromedioDiasResolucion,
	)

	if err != nil {
		log.Printf("⚠️ Error obteniendo estadísticas: %v", err) // Log para debug
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error obteniendo estadísticas"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_reclamos":           stats.TotalReclamos,
			"pendientes":               stats.Pendientes,
			"en_proceso":               stats.EnProceso,
			"resueltos":                stats.Resueltos,
			"cerrados":                 stats.Cerrados,
			"reclamos_hoy":             stats.ReclamosHoy,
			"reclamos_semana":          stats.ReclamosSemana,
			"reclamos_mes":             stats.ReclamosMes,
			"promedio_dias_resolucion": stats.PromedioDiasResolucion,
		},
	})
}