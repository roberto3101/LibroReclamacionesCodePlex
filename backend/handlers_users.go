package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Estructura para listar
type UsuarioResponse struct {
	ID            string     `json:"id"`
	Email         string     `json:"email"`
	Nombre        string     `json:"nombre_completo"`
	Rol           string     `json:"rol"`
	Activo        bool       `json:"activo"`
	UltimoAcceso  *time.Time `json:"ultimo_acceso"`
	FechaCreacion time.Time  `json:"fecha_creacion"`
}

// GET /api/admin/usuarios
func listarUsuariosHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	rows, err := pool.Query(ctx, `
		SELECT id, email, nombre_completo, rol, activo, ultimo_acceso, fecha_creacion 
		FROM usuarios_admin ORDER BY fecha_creacion DESC
	`)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": "Error consultando usuarios"})
		return
	}
	defer rows.Close()

	var usuarios []UsuarioResponse
	for rows.Next() {
		var u UsuarioResponse
		if err := rows.Scan(&u.ID, &u.Email, &u.Nombre, &u.Rol, &u.Activo, &u.UltimoAcceso, &u.FechaCreacion); err == nil {
			usuarios = append(usuarios, u)
		}
	}
	c.JSON(200, gin.H{"success": true, "data": usuarios})
}

// POST /api/admin/usuarios
// POST /api/admin/usuarios
func crearUsuarioHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Nombre   string `json:"nombre_completo" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
		Rol      string `json:"rol" binding:"required,oneof=ADMIN SOPORTE"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "message": "Datos inválidos: " + err.Error()})
		return
	}

	// 1. Encriptar contraseña (HASHING)
	log.Printf("🔒 Encriptando contraseña para usuario: %s", req.Email)
	hash, errHash := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if errHash != nil {
		log.Printf("❌ Error generando hash: %v", errHash)
		c.JSON(500, gin.H{"success": false, "message": "Error interno de seguridad"})
		return
	}

	// Recuperar ID del admin que lo crea (si existe)
	creadoPorID, _ := c.Get("user_id")

	// 2. Insertar en BD el HASH (string(hash)), NO la contraseña plana
	var nuevoID string
	query := `
		INSERT INTO usuarios_admin (email, nombre_completo, password_hash, rol, creado_por)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	err := pool.QueryRow(ctx, query, req.Email, req.Nombre, string(hash), req.Rol, creadoPorID).Scan(&nuevoID)

	if err != nil {
		log.Printf("❌ Error en BD al crear usuario: %v", err)
		c.JSON(500, gin.H{"success": false, "message": "Error al guardar usuario (¿Email duplicado?)"})
		return
	}

	log.Printf("✅ Usuario creado con éxito. ID: %s", nuevoID)
	c.JSON(201, gin.H{"success": true, "message": "Usuario creado correctamente"})
}

// PUT /api/admin/usuarios/:id (Editar rol, nombre, activo)
func editarUsuarioHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id := c.Param("id")
	var req struct {
		Nombre string `json:"nombre_completo"`
		Rol    string `json:"rol"`
		Activo *bool  `json:"activo"` // Puntero para detectar false
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "message": "Datos inválidos"})
		return
	}

	// Construcción dinámica de query (simple)
	_, err := pool.Exec(ctx, `
		UPDATE usuarios_admin 
		SET nombre_completo = COALESCE(NULLIF($1, ''), nombre_completo),
		    rol = COALESCE(NULLIF($2, ''), rol),
		    activo = COALESCE($3, activo)
		WHERE id = $4
	`, req.Nombre, req.Rol, req.Activo, id)

	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": "Error al actualizar"})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "Usuario actualizado"})
}

// PUT /api/admin/usuarios/:id/password (Admin cambia password a otro)
func cambiarPasswordAdminHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	id := c.Param("id")
	var req struct {
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "message": "Contraseña muy corta"})
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)

	_, err := pool.Exec(ctx, "UPDATE usuarios_admin SET password_hash = $1 WHERE id = $2", string(hash), id)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": "Error actualizando contraseña"})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "Contraseña actualizada"})
}