// =============================================================================
// API BACKEND - LIBRO DE RECLAMACIONES CODEPLEX
// Go + Gin + PostgreSQL/CockroachDB
// Equivalente directo a server.js
// =============================================================================

package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"sync"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
)

// =============================================================================
// VARIABLES GLOBALES
// =============================================================================

var (
	pool   *pgxpool.Pool
	dialer *gomail.Dialer
	config Config
)

// =============================================================================
// CONFIGURACIÓN
// =============================================================================

type Config struct {
	Port         string
	Env          string
	DatabaseURL  string
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPass     string
	SMTPFrom     string
	FrontendURL  string
	BackendURL   string
	EmailSoporte string
}

func loadConfig() Config {
	env := os.Getenv("NODE_ENV")
	if env == "" {
		env = "development"
	}

	envFile := ".env"
	if env == "production" {
		envFile = ".env.production"
	}
	_ = godotenv.Load(envFile)

	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	return Config{
		Port:         getEnv("PORT", "3000"),
		Env:          env,
		DatabaseURL:  getEnv("DATABASE_URL", "postgresql://postgres:sql@127.0.0.1:5432/libro_reclamaciones"),
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     smtpPort,
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPass:     getEnv("SMTP_PASS", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "libro.reclamaciones@codeplex.pe"),
		FrontendURL:  getEnv("FRONTEND_URL", "http://localhost:4321"),
		BackendURL:   getEnv("BACKEND_URL", "http://localhost:3000"),
		EmailSoporte: getEnv("EMAIL_SOPORTE", "soporte@codeplex.pe"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// =============================================================================
// ESTRUCTURAS
// =============================================================================

type CrearReclamoRequest struct {
	TipoSolicitud        string   `json:"tipo_solicitud"`
	NombreCompleto       string   `json:"nombre_completo"`
	TipoDocumento        string   `json:"tipo_documento"`
	NumeroDocumento      string   `json:"numero_documento"`
	Telefono             string   `json:"telefono"`
	Email                string   `json:"email"`
	Domicilio            *string  `json:"domicilio"`
	Departamento         *string  `json:"departamento"`
	Provincia            *string  `json:"provincia"`
	Distrito             *string  `json:"distrito"`
	TipoBien             *string  `json:"tipo_bien"`
	MontoReclamado       float64  `json:"monto_reclamado"`
	DescripcionBien      string   `json:"descripcion_bien"`
	AreaQueja            *string  `json:"area_queja"`
	DescripcionSituacion *string  `json:"descripcion_situacion"`
	FechaIncidente       string   `json:"fecha_incidente"`
	DetalleReclamo       string   `json:"detalle_reclamo"`
	PedidoConsumidor     string   `json:"pedido_consumidor"`
	FirmaDigital         string   `json:"firma_digital"`
	AceptaTerminos       bool     `json:"acepta_terminos"`
	AceptaCopia          bool     `json:"acepta_copia"`
}

type ReclamoCreado struct {
	ID                   string    `json:"id"`
	CodigoReclamo        string    `json:"codigo_reclamo"`
	FechaRegistro        time.Time `json:"fecha_registro"`
	FechaLimiteRespuesta time.Time `json:"fecha_limite_respuesta"`
}

// =============================================================================
// FUNCIÓN: Generar código único
// =============================================================================

func generarCodigoReclamo(ctx context.Context) (string, error) {
	año := time.Now().Year()
	pattern := fmt.Sprintf("CODEPLEX-%d-%%", año)

	var ultimoCodigo sql.NullString
	err := pool.QueryRow(ctx, `
		SELECT codigo_reclamo FROM reclamos 
		WHERE codigo_reclamo LIKE $1 
		ORDER BY codigo_reclamo DESC LIMIT 1
	`, pattern).Scan(&ultimoCodigo)

	if err != nil && err.Error() != "no rows in result set" {
		return "", err
	}

	numero := 1
	if ultimoCodigo.Valid {
		parts := strings.Split(ultimoCodigo.String, "-")
		if len(parts) == 3 {
			if n, err := strconv.Atoi(parts[2]); err == nil {
				numero = n + 1
			}
		}
	}

	return fmt.Sprintf("CODEPLEX-%d-%05d", año, numero), nil
}

// =============================================================================
// HANDLERS
// =============================================================================

// Health check
func healthHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	dbStatus := "disconnected"
	if err := pool.Ping(ctx); err == nil {
		dbStatus = "connected"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"database":  dbStatus,
	})
}

// POST /api/reclamos - Crear nuevo reclamo
func crearReclamoHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	var req CrearReclamoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Datos inválidos: " + err.Error(),
		})
		return
	}

	// Validaciones (igual que en Node.js)
	if req.TipoSolicitud != "RECLAMO" && req.TipoSolicitud != "QUEJA" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Tipo de solicitud inválido"})
		return
	}

	if req.FirmaDigital == "" || !strings.HasPrefix(req.FirmaDigital, "data:image") {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Firma digital requerida"})
		return
	}

	if !req.AceptaTerminos {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Debe aceptar los términos y condiciones"})
		return
	}

	emailRegex := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	if !emailRegex.MatchString(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Formato de correo electrónico inválido"})
		return
	}

	if req.DescripcionBien == "" || req.DetalleReclamo == "" || req.PedidoConsumidor == "" {
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Faltan detalles del reclamo o el pedido del consumidor"})
        return
    }

    // CORRECCIÓN: Validar que el monto no exceda el límite de DECIMAL(10,2)
    if req.MontoReclamado > 9999999.99 {
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "El monto reclamado excede el límite permitido"})
        return
    }

	if len(req.DetalleReclamo) > 3000 || len(req.PedidoConsumidor) > 2000 || len(req.NombreCompleto) > 200 || len(req.DescripcionBien) > 600 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Uno de los campos excede el límite permitido de caracteres."})
		return
	}

	// Iniciar transacción
	tx, err := pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Error al iniciar transacción"})
		return
	}
	defer tx.Rollback(ctx)

	// Generar código único
	codigoReclamo, err := generarCodigoReclamo(ctx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Error generando código"})
		return
	}

	// Insertar reclamo
	var reclamo ReclamoCreado
	err = tx.QueryRow(ctx, `
		INSERT INTO reclamos (
			codigo_reclamo, tipo_solicitud, nombre_completo, tipo_documento, numero_documento,
			telefono, email, domicilio, departamento, provincia, distrito,
			tipo_bien, monto_reclamado, descripcion_bien, area_queja, descripcion_situacion,
			fecha_incidente, detalle_reclamo, pedido_consumidor, firma_digital,
			acepta_terminos, acepta_copia, ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)
		RETURNING id, codigo_reclamo, fecha_registro, fecha_limite_respuesta
	`,
		codigoReclamo, req.TipoSolicitud, req.NombreCompleto, req.TipoDocumento, req.NumeroDocumento,
		req.Telefono, req.Email, nullString(req.Domicilio), nullString(req.Departamento), nullString(req.Provincia), nullString(req.Distrito),
		nullString(req.TipoBien), req.MontoReclamado, req.DescripcionBien, nullString(req.AreaQueja), nullString(req.DescripcionSituacion),
		req.FechaIncidente, req.DetalleReclamo, req.PedidoConsumidor, req.FirmaDigital,
		req.AceptaTerminos, req.AceptaCopia, c.ClientIP(), c.GetHeader("User-Agent"),
	).Scan(&reclamo.ID, &reclamo.CodigoReclamo, &reclamo.FechaRegistro, &reclamo.FechaLimiteRespuesta)

	if err != nil {
		log.Printf("Error insertando reclamo: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Error al registrar el reclamo",
			"error":   errorDetail(err),
		})
		return
	}

	if err := tx.Commit(ctx); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Error al confirmar transacción"})
        return
    }

  // Unificamos el envío de emails de forma segura después del commit
    go func() {
        // Determinamos el tipo de bien (Producto/Servicio)
        finalTipoBien := "SERVICIO"
        if req.TipoBien != nil && *req.TipoBien != "" {
            finalTipoBien = *req.TipoBien
        }
        
        if err := enviarEmails(reclamo, req, finalTipoBien); err != nil {
            log.Printf("Error enviando emails: %v", err)
        }
    }()

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Reclamo registrado exitosamente",
		"data": gin.H{
			"codigo_reclamo":         reclamo.CodigoReclamo,
			"fecha_registro":         reclamo.FechaRegistro,
			"fecha_limite_respuesta": reclamo.FechaLimiteRespuesta,
			"plazo_dias":             15,
		},
	})
}

// GET /api/reclamos/:codigo - Consultar reclamo por código
func obtenerReclamoHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	codigo := c.Param("codigo")

	rows, err := pool.Query(ctx, `
		SELECT 
			r.id, r.codigo_reclamo, r.tipo_solicitud, r.estado,
			r.nombre_completo, r.tipo_documento, r.numero_documento,
			r.telefono, r.email, r.domicilio, r.departamento, r.provincia, r.distrito,
			r.razon_social, r.ruc, r.direccion_proveedor,
			r.tipo_bien, r.monto_reclamado, r.descripcion_bien,
			r.area_queja, r.descripcion_situacion,
			r.fecha_incidente, r.detalle_reclamo, r.pedido_consumidor,
			r.acepta_terminos, r.acepta_copia,
			r.fecha_registro, r.fecha_limite_respuesta, r.fecha_respuesta,
			res.respuesta_empresa, res.fecha_respuesta as res_fecha, res.respondido_por
		FROM reclamos r
		LEFT JOIN respuestas res ON r.id = res.reclamo_id
		WHERE r.codigo_reclamo = $1
	`, codigo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error al consultar el reclamo"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Reclamo no encontrado"})
		return
	}

	var (
		id, codigoReclamo, tipoSolicitud, estado                        string
		nombreCompleto, tipoDocumento, numeroDocumento, telefono, email string
		domicilio, departamento, provincia, distrito                    sql.NullString
		razonSocial, ruc, direccionProveedor                            string
		tipoBien, areaQueja, descripcionSituacion                       sql.NullString
		montoReclamado                                                  float64
		descripcionBien, detalleReclamo, pedidoConsumidor               string
		fechaIncidente                                                  time.Time
		aceptaTerminos, aceptaCopia                                     bool
		fechaRegistro, fechaLimiteRespuesta                             time.Time
		fechaRespuesta                                                  sql.NullTime
		respuestaEmpresa, resFecha, respondidoPor                       sql.NullString
	)

	err = rows.Scan(
		&id, &codigoReclamo, &tipoSolicitud, &estado,
		&nombreCompleto, &tipoDocumento, &numeroDocumento, &telefono, &email,
		&domicilio, &departamento, &provincia, &distrito,
		&razonSocial, &ruc, &direccionProveedor,
		&tipoBien, &montoReclamado, &descripcionBien,
		&areaQueja, &descripcionSituacion,
		&fechaIncidente, &detalleReclamo, &pedidoConsumidor,
		&aceptaTerminos, &aceptaCopia,
		&fechaRegistro, &fechaLimiteRespuesta, &fechaRespuesta,
		&respuestaEmpresa, &resFecha, &respondidoPor,
	)
	if err != nil {
		log.Printf("Error escaneando reclamo: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error al procesar el reclamo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":                     id,
			"codigo_reclamo":         codigoReclamo,
			"tipo_solicitud":         tipoSolicitud,
			"estado":                 estado,
			"nombre_completo":        nombreCompleto,
			"tipo_documento":         tipoDocumento,
			"numero_documento":       numeroDocumento,
			"telefono":               telefono,
			"email":                  email,
			"domicilio":              nullToInterface(domicilio),
			"departamento":           nullToInterface(departamento),
			"provincia":              nullToInterface(provincia),
			"distrito":               nullToInterface(distrito),
			"razon_social":           razonSocial,
			"ruc":                    ruc,
			"direccion_proveedor":    direccionProveedor,
			"tipo_bien":              nullToInterface(tipoBien),
			"monto_reclamado":        montoReclamado,
			"descripcion_bien":       descripcionBien,
			"area_queja":             nullToInterface(areaQueja),
			"descripcion_situacion":  nullToInterface(descripcionSituacion),
			"fecha_incidente":        fechaIncidente,
			"detalle_reclamo":        detalleReclamo,
			"pedido_consumidor":      pedidoConsumidor,
			"acepta_terminos":        aceptaTerminos,
			"acepta_copia":           aceptaCopia,
			"fecha_registro":         fechaRegistro,
			"fecha_limite_respuesta": fechaLimiteRespuesta,
			"fecha_respuesta":        nullTimeToInterface(fechaRespuesta),
			"respuesta_empresa":      nullToInterface(respuestaEmpresa),
			"respondido_por":         nullToInterface(respondidoPor),
		},
	})
}

// GET /api/reclamos/:codigo/firma - Ver firma digital
func obtenerFirmaHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	codigo := c.Param("codigo")

	var firma string
	err := pool.QueryRow(ctx, "SELECT firma_digital FROM reclamos WHERE codigo_reclamo = $1", codigo).Scan(&firma)
	if err != nil {
		c.String(http.StatusNotFound, "Reclamo no encontrado")
		return
	}

	parts := strings.SplitN(firma, ",", 2)
	if len(parts) != 2 {
		c.String(http.StatusInternalServerError, "Formato de firma inválido")
		return
	}

	imageData, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		c.String(http.StatusInternalServerError, "Error decodificando firma")
		return
	}

	c.Header("Content-Type", "image/png")
	c.Data(http.StatusOK, "image/png", imageData)
}

// GET /api/dashboard - Dashboard de reclamos
func dashboardHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var pendientes, enProceso, resueltos, vencidos, totalReclamos, totalQuejas, total int64
	err := pool.QueryRow(ctx, "SELECT * FROM dashboard_reclamos").Scan(
		&pendientes, &enProceso, &resueltos, &vencidos, &totalReclamos, &totalQuejas, &total,
	)
	if err != nil {
		log.Printf("Error en dashboard stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error al obtener dashboard"})
		return
	}

	rows, err := pool.Query(ctx, "SELECT * FROM reclamos_pendientes LIMIT 10")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Error al obtener dashboard"})
		return
	}
	defer rows.Close()

	var pendientesList []gin.H
	for rows.Next() {
		var id, codigoReclamo, tipoSolicitud, nombreCompleto, email string
		var fechaRegistro, fechaLimiteRespuesta time.Time
		var diasRestantes int
		var prioridad string

		if err := rows.Scan(&id, &codigoReclamo, &tipoSolicitud, &nombreCompleto, &email, &fechaRegistro, &fechaLimiteRespuesta, &diasRestantes, &prioridad); err != nil {
			continue
		}

		pendientesList = append(pendientesList, gin.H{
			"id":                     id,
			"codigo_reclamo":         codigoReclamo,
			"tipo_solicitud":         tipoSolicitud,
			"nombre_completo":        nombreCompleto,
			"email":                  email,
			"fecha_registro":         fechaRegistro,
			"fecha_limite_respuesta": fechaLimiteRespuesta,
			"dias_restantes":         diasRestantes,
			"prioridad":              prioridad,
		})
	}

	if pendientesList == nil {
		pendientesList = []gin.H{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"estadisticas": gin.H{
				"pendientes":     pendientes,
				"en_proceso":     enProceso,
				"resueltos":      resueltos,
				"vencidos":       vencidos,
				"total_reclamos": totalReclamos,
				"total_quejas":   totalQuejas,
				"total":          total,
			},
			"pendientes": pendientesList,
		},
	})
}

// =============================================================================
// ENVÍO DE EMAILS
// =============================================================================

func enviarEmails(reclamo ReclamoCreado, req CrearReclamoRequest, tipoBien string) error {
	if reclamo.CodigoReclamo == "" {
		return fmt.Errorf("omitido: reclamo sin codigo")
	}

	fechaLimite := reclamo.FechaLimiteRespuesta.Format("02/01/2006")
	fechaRegistro := reclamo.FechaRegistro.Format("02/01/2006 15:04:05")

	val := func(s *string) string {
		if s == nil { return "" }
		return *s
	}

	ubicacion := "No especificada"
	dep, prov, dist, dom := val(req.Departamento), val(req.Provincia), val(req.Distrito), val(req.Domicilio)
	if dep != "" || prov != "" || dist != "" {
		ubicacion = fmt.Sprintf("%s / %s - %s", dep, prov, dist)
		if dom != "" { ubicacion = dom + " (" + ubicacion + ")" }
	}

	// Canal para esperar ambos envíos
	var wg sync.WaitGroup

	// Email a soporte (en paralelo)
	wg.Add(1)
	go func() {
		defer wg.Done()
		m := gomail.NewMessage()
		m.SetHeader("From", config.SMTPFrom)
		m.SetHeader("To", config.EmailSoporte)
		m.SetHeader("Subject", fmt.Sprintf("Nuevo %s - %s", req.TipoSolicitud, reclamo.CodigoReclamo))
		m.SetBody("text/html", generarEmailSoporte(reclamo.CodigoReclamo, req.TipoSolicitud, fechaLimite, req.NombreCompleto, req.Email, req.Telefono, ubicacion, tipoBien, req.DescripcionBien, req.DetalleReclamo, req.MontoReclamado))
		
		if err := dialer.DialAndSend(m); err != nil {
			log.Printf("❌ Error email soporte: %v", err)
		} else {
			log.Printf("✅ Email enviado a soporte: %s", config.EmailSoporte)
		}
	}()

	// Email al cliente (en paralelo)
	if req.AceptaCopia && req.Email != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m2 := gomail.NewMessage()
			m2.SetHeader("From", config.SMTPFrom)
			m2.SetHeader("To", req.Email)
			m2.SetHeader("Subject", fmt.Sprintf("Confirmación de %s - %s", req.TipoSolicitud, reclamo.CodigoReclamo))
			m2.SetBody("text/html", generarEmailCliente(reclamo.CodigoReclamo, req.TipoSolicitud, fechaLimite, fechaRegistro, req.NombreCompleto, tipoBien, req.DescripcionBien))
			
			if err := dialer.DialAndSend(m2); err != nil {
				log.Printf("❌ Error email cliente: %v", err)
			} else {
				log.Printf("✅ Email enviado a cliente: %s", req.Email)
			}
		}()
	}

	// Esperar que ambos terminen (con timeout)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("📧 Todos los emails procesados para %s", reclamo.CodigoReclamo)
	case <-time.After(30 * time.Second):
		log.Printf("⚠️ Timeout enviando emails para %s", reclamo.CodigoReclamo)
	}

	return nil
}

// Se agregaron argumentos: telefono, ubicacion
func generarEmailSoporte(codigo, tipo, fechaLimite, nombre, email, telefono, ubicacion, tipoBien, descripcion, detalle string, monto float64) string {
	tipoBienBg := "#ffedd5"
	tipoBienColor := "#9a3412"
	if tipoBien == "PRODUCTO" {
		tipoBienBg = "#dbeafe"
		tipoBienColor = "#1e40af"
	}

	montoHTML := ""
	if monto > 0 {
		montoHTML = fmt.Sprintf(`<div style="margin-top: 10px; padding: 8px; background-color: #fef3c7; border-radius: 4px;"><strong>💰 Monto Reclamado:</strong> S/ %.2f</div>`, monto)
	}

	pedidoHTML := ""
	if detalle != "" {
		pedidoHTML = detalle
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html><head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Nuevo %s - %s</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Arial, sans-serif; font-size: 14px; line-height: 1.5; color: #333333; background-color: #f3f4f6; -webkit-text-size-adjust: 100%%; -ms-text-size-adjust: 100%%;">

<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background-color: #f3f4f6;">
<tr><td align="center" style="padding: 20px 10px;">

<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="max-width: 600px; background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">

<!-- Header -->
<tr>
<td style="background: #1e40af; padding: 25px 20px; text-align: center;">
<h1 style="margin: 0; color: #ffffff; font-size: 22px; font-weight: bold;">📋 Nuevo %s</h1>
<p style="margin: 8px 0 0 0; color: #93c5fd; font-size: 15px;">Código: <strong style="color: #ffffff;">%s</strong></p>
</td>
</tr>

<!-- Alerta Plazo -->
<tr>
<td style="padding: 20px 20px 15px 20px;">
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background-color: #fef2f2; border-left: 4px solid #dc2626; border-radius: 0 4px 4px 0;">
<tr>
<td style="padding: 12px 15px;">
<strong style="color: #991b1b;">⚠️ PLAZO LEGAL:</strong><br>
<span style="color: #7f1d1d;">Debe responder antes del <strong>%s</strong> (15 días hábiles)</span>
</td>
</tr>
</table>
</td>
</tr>

<!-- Datos del Consumidor -->
<tr>
<td style="padding: 0 20px 15px 20px;">
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background-color: #eff6ff; border-radius: 6px; border: 1px solid #bfdbfe;">
<tr>
<td style="padding: 10px 15px; border-bottom: 1px solid #bfdbfe; font-weight: bold; color: #1e40af; font-size: 14px;">
👤 Datos del Consumidor
</td>
</tr>
<tr>
<td style="padding: 15px;">
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
<tr><td style="padding: 5px 0; word-break: break-word;"><strong>Nombre:</strong> %s</td></tr>
<tr><td style="padding: 5px 0; word-break: break-all;"><strong>Email:</strong> <a href="mailto:%s" style="color: #1e40af; text-decoration: none;">%s</a></td></tr>
<tr><td style="padding: 5px 0;"><strong>Teléfono:</strong> <a href="tel:%s" style="color: #1e40af; text-decoration: none; font-weight: bold;">%s</a></td></tr>
<tr><td style="padding: 5px 0; word-break: break-word;"><strong>Ubicación:</strong> %s</td></tr>
</table>
</td>
</tr>
</table>
</td>
</tr>

<!-- Bien Contratado -->
<tr>
<td style="padding: 0 20px 15px 20px;">
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background-color: #ffffff; border-radius: 6px; border: 1px solid #e5e7eb;">
<tr>
<td style="padding: 10px 15px; border-bottom: 1px solid #e5e7eb; font-weight: bold; color: #1e40af; font-size: 14px;">
📦 Bien Contratado 
<span style="background-color: %s; color: %s; padding: 2px 8px; border-radius: 10px; font-size: 11px; margin-left: 5px;">%s</span>
</td>
</tr>
<tr>
<td style="padding: 15px;">
<div style="word-break: break-word; overflow-wrap: break-word; white-space: pre-line;">%s</div>
%s
</td>
</tr>
</table>
</td>
</tr>

<!-- Detalle del Reclamo -->
<tr>
<td style="padding: 0 20px 15px 20px;">
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background-color: #fff7ed; border-radius: 6px; border: 1px solid #fed7aa;">
<tr>
<td style="padding: 10px 15px; border-bottom: 1px solid #fed7aa; font-weight: bold; color: #c2410c; font-size: 14px;">
📝 Detalle del %s
</td>
</tr>
<tr>
<td style="padding: 15px;">
<div style="word-break: break-word; overflow-wrap: break-word; white-space: pre-line; color: #1f2937;">%s</div>
</td>
</tr>
</table>
</td>
</tr>

<!-- Firma Digital -->
<tr>
<td style="padding: 0 20px 20px 20px;">
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background-color: #f0fdf4; border-radius: 6px; border: 1px solid #86efac;">
<tr>
<td style="padding: 15px; text-align: center;">
<span style="color: #166534; font-weight: bold;">🖊️ Firma Digital:</span>
<a href="%s/api/reclamos/%s/firma" target="_blank" style="display: inline-block; margin-left: 8px; padding: 8px 16px; background-color: #1e40af; color: #ffffff; text-decoration: none; border-radius: 4px; font-weight: bold; font-size: 13px;">Ver Firma</a>
</td>
</tr>
</table>
</td>
</tr>

<!-- Footer -->
<tr>
<td style="background-color: #1f2937; padding: 20px; text-align: center;">
<p style="margin: 0; color: #9ca3af; font-size: 11px;">Este email fue generado automáticamente por el Libro de Reclamaciones Virtual</p>
<p style="margin: 6px 0 0 0; color: #6b7280; font-size: 10px;">CODEPLEX SAC | RUC: 20539782232</p>
</td>
</tr>

</table>
</td></tr>
</table>
</body></html>`,
		tipo, codigo,
		tipo, codigo,
		fechaLimite,
		nombre, email, email, telefono, telefono, ubicacion,
		tipoBienBg, tipoBienColor, tipoBien, descripcion, montoHTML,
		tipo, pedidoHTML,
		config.BackendURL, codigo)
}

func generarEmailCliente(codigo, tipo, fechaLimite, fechaRegistro, nombre, tipoBien, descripcion string) string {
	tipoLower := "reclamo"
	if tipo == "QUEJA" {
		tipoLower = "queja"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html><head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Confirmación de %s - %s</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Arial, sans-serif; font-size: 14px; line-height: 1.5; color: #333333; background-color: #f3f4f6; -webkit-text-size-adjust: 100%%; -ms-text-size-adjust: 100%%;">

<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background-color: #f3f4f6;">
<tr><td align="center" style="padding: 20px 10px;">

<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="max-width: 600px; background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">

<!-- Header Verde -->
<tr>
<td style="background: #059669; padding: 25px 20px; text-align: center;">
<h1 style="margin: 0; color: #ffffff; font-size: 22px; font-weight: bold;">✅ %s Registrado</h1>
<p style="margin: 8px 0 0 0; color: #a7f3d0; font-size: 16px; font-weight: bold;">%s</p>
</td>
</tr>

<!-- Mensaje de Confirmación -->
<tr>
<td style="padding: 20px 20px 15px 20px;">
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background-color: #f0fdf4; border-left: 4px solid #10b981; border-radius: 0 4px 4px 0;">
<tr>
<td style="padding: 15px;">
<strong style="color: #166534;">Estimado/a %s</strong><br>
<span style="color: #15803d;">Su %s ha sido registrado exitosamente en nuestro Libro de Reclamaciones Virtual.</span>
</td>
</tr>
</table>
</td>
</tr>

<!-- Plazo de Respuesta -->
<tr>
<td style="padding: 0 20px 15px 20px;">
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background-color: #eff6ff; border-left: 4px solid #3b82f6; border-radius: 0 4px 4px 0;">
<tr>
<td style="padding: 15px;">
<strong style="color: #1e40af;">📅 Plazo de Respuesta:</strong><br>
<span style="color: #1e3a8a;">Recibirá nuestra respuesta antes del <strong>%s</strong></span><br>
<span style="color: #6b7280; font-size: 12px;">Plazo legal: 15 días hábiles (según D.S. 011-2011-PCM)</span>
</td>
</tr>
</table>
</td>
</tr>

<!-- Resumen -->
<tr>
<td style="padding: 0 20px 15px 20px;">
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background-color: #ffffff; border-radius: 6px; border: 1px solid #e5e7eb;">
<tr>
<td style="padding: 10px 15px; border-bottom: 1px solid #e5e7eb; font-weight: bold; color: #374151; background-color: #f9fafb; font-size: 14px;">
📋 Resumen de su %s
</td>
</tr>
<tr>
<td style="padding: 0;">
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
<tr>
<td style="padding: 10px 15px; border-bottom: 1px solid #f3f4f6; width: 120px; font-weight: bold; color: #6b7280;">Código:</td>
<td style="padding: 10px 15px; border-bottom: 1px solid #f3f4f6; color: #1e40af; font-weight: bold;">%s</td>
</tr>
<tr>
<td style="padding: 10px 15px; border-bottom: 1px solid #f3f4f6; font-weight: bold; color: #6b7280; background-color: #f9fafb;">Fecha:</td>
<td style="padding: 10px 15px; border-bottom: 1px solid #f3f4f6; background-color: #f9fafb;">%s</td>
</tr>
<tr>
<td style="padding: 10px 15px; border-bottom: 1px solid #f3f4f6; font-weight: bold; color: #6b7280;">Tipo:</td>
<td style="padding: 10px 15px; border-bottom: 1px solid #f3f4f6;">%s</td>
</tr>
<tr>
<td style="padding: 10px 15px; border-bottom: 1px solid #f3f4f6; font-weight: bold; color: #6b7280; background-color: #f9fafb;">Bien:</td>
<td style="padding: 10px 15px; border-bottom: 1px solid #f3f4f6; background-color: #f9fafb;">%s</td>
</tr>
<tr>
<td style="padding: 10px 15px; font-weight: bold; color: #6b7280; vertical-align: top;">Descripción:</td>
<td style="padding: 10px 15px; word-break: break-word; overflow-wrap: break-word; white-space: pre-line;">%s</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>

<!-- Información Importante -->
<tr>
<td style="padding: 0 20px 20px 20px;">
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background-color: #fefce8; border-left: 4px solid #eab308; border-radius: 0 4px 4px 0;">
<tr>
<td style="padding: 15px;">
<strong style="color: #a16207;">ℹ️ Información Importante:</strong><br>
<span style="color: #854d0e;">• Conserve este email como comprobante de su solicitud</span><br>
<span style="color: #854d0e;">• Le notificaremos cuando tengamos una respuesta</span>
</td>
</tr>
</table>
</td>
</tr>

<!-- Footer -->
<tr>
<td style="background-color: #1f2937; padding: 20px; text-align: center;">
<p style="margin: 0; color: #ffffff; font-weight: bold; font-size: 13px;">CODEPLEX SAC</p>
<p style="margin: 4px 0; color: #9ca3af; font-size: 11px;">RUC: 20539782232</p>
<p style="margin: 8px 0 0 0; color: #6b7280; font-size: 10px;">AV. LOS PROCERES MZA. G3 LOTE. 11 - LIMA - LOS OLIVOS</p>
<p style="margin: 4px 0 0 0; color: #6b7280; font-size: 10px;">
<a href="mailto:soporte@codeplex.pe" style="color: #60a5fa; text-decoration: none;">soporte@codeplex.pe</a> | 
<a href="tel:+51936343607" style="color: #60a5fa; text-decoration: none;">+51 936343607</a>
</p>
</td>
</tr>

</table>
</td></tr>
</table>
</body></html>`,
		tipo, codigo,
		tipo, codigo,
		nombre, tipoLower,
		fechaLimite,
		tipo, codigo, fechaRegistro, tipo, tipoBien, descripcion)
}
// =============================================================================
// HELPERS
// =============================================================================

func nullString(s *string) sql.NullString {
	// CORRECCIÓN: Validación explícita de nil para evitar Puntero Nulo (Panic)
	if s == nil {
		return sql.NullString{Valid: false}
	}
	if *s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}


func nullToInterface(ns sql.NullString) interface{} {
	if ns.Valid {
		return ns.String
	}
	return nil
}

func nullTimeToInterface(nt sql.NullTime) interface{} {
	if nt.Valid {
		return nt.Time
	}
	return nil
}

func errorDetail(err error) string {
	if config.Env == "development" {
		return err.Error()
	}
	return ""
}

// =============================================================================
// MIDDLEWARE CORS
// =============================================================================

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// =============================================================================
// MAIN
// =============================================================================

func main() {
	config = loadConfig()

	if config.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Conectar a la base de datos
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	pool, err = pgxpool.New(ctx, config.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Error conectando a la base de datos: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("❌ Error verificando conexión: %v", err)
	}
	log.Println("✅ Conectado a la base de datos")

	// Configurar SMTP
	dialer = gomail.NewDialer(config.SMTPHost, config.SMTPPort, config.SMTPUser, config.SMTPPass)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true, ServerName: config.SMTPHost}
	if config.SMTPPort == 465 {
		dialer.SSL = true
	}

// Router
	router := gin.New()
	
	// CORRECCIÓN: Configuración para Proxies (Vercel/Cloudflare/Nginx)
	router.ForwardedByClientIP = true
	router.SetTrustedProxies(nil) // Confía en todos los proxies de entrada (necesario para la nube)

	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(corsMiddleware())

	// Rutas
	api := router.Group("/api")
	{
		api.GET("/health", healthHandler)
		api.POST("/reclamos", crearReclamoHandler)
		api.GET("/reclamos/:codigo", obtenerReclamoHandler)
		api.GET("/reclamos/:codigo/firma", obtenerFirmaHandler)
		api.GET("/dashboard", dashboardHandler)
	}

	// Servidor
	srv := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("🚀 Servidor corriendo en puerto %s", config.Port)
		log.Printf("📍 Health check: http://localhost:%s/api/health", config.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("👋 Cerrando servidor...")
	ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel2()
	srv.Shutdown(ctx2)
}