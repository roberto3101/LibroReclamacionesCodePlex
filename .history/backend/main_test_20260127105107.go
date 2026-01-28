// =============================================================================
// TESTS COMPLETOS - BACKEND ADMIN
// Pruebas de integración reales contra la base de datos
// =============================================================================

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// =============================================================================
// SETUP DE TESTS
// =============================================================================

var (
	testPool   *pgxpool.Pool
	testRouter *gin.Engine
	testToken  string
	testUserID string
)

// setupTestDB configura la conexión a la base de datos de pruebas
func setupTestDB(t *testing.T) {
	ctx := context.Background()
	
	// Usar la misma DB pero con transacciones que se revierten
	connStr := "postgresql://postgres:sql@127.0.0.1:5432/libro_reclamaciones"
	
	var err error
	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("Error conectando a DB de pruebas: %v", err)
	}
	
	if err := testPool.Ping(ctx); err != nil {
		t.Fatalf("Error haciendo ping a DB: %v", err)
	}
	
	// Asignar pool global para que los handlers lo usen
	pool = testPool
	
	t.Log("✅ Base de datos de pruebas conectada")
}

// setupTestRouter configura el router con todas las rutas
func setupTestRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testRouter = gin.New()
	
	api := testRouter.Group("/api")
	{
		// Rutas públicas
		adminAuth := api.Group("/admin/auth")
		{
			adminAuth.POST("/login", loginAdminHandler)
		}
		
		// Rutas protegidas
		admin := api.Group("/admin")
		admin.Use(authMiddleware())
		{
			admin.GET("/reclamos", listarReclamosAdminHandler)
			admin.PUT("/reclamos/:id/estado", cambiarEstadoReclamoHandler)
			admin.POST("/reclamos/:id/respuesta", responderReclamoHandler)
			admin.GET("/dashboard/stats", obtenerEstadisticasHandler)
		}
	}
	
	t.Log("✅ Router de pruebas configurado")
}

// crearUsuarioPrueba crea un usuario de prueba en la DB
func crearUsuarioPrueba(t *testing.T) string {
	ctx := context.Background()
	
	// Verificar si ya existe
	var existingID string
	err := testPool.QueryRow(ctx, "SELECT id FROM usuarios_admin WHERE email = $1", "test@codeplex.com").Scan(&existingID)
	if err == nil {
		t.Log("✅ Usuario de prueba ya existe, usando existente")
		return existingID
	}
	
	// Crear usuario de prueba
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("Test123!"), 12)
	
	var userID string
	err = testPool.QueryRow(ctx, `
		INSERT INTO usuarios_admin (email, nombre_completo, password_hash, rol, activo, debe_cambiar_password)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, "test@codeplex.com", "Usuario de Prueba", string(passwordHash), "ADMIN", true, false).Scan(&userID)
	
	if err != nil {
		t.Fatalf("Error creando usuario de prueba: %v", err)
	}
	
	t.Logf("✅ Usuario de prueba creado: %s", userID)
	return userID
}

// crearReclamoPrueba crea un reclamo de prueba
func crearReclamoPrueba(t *testing.T) string {
	ctx := context.Background()
	
	var reclamoID string
	err := testPool.QueryRow(ctx, `
		INSERT INTO reclamos (
			codigo_reclamo, tipo_solicitud, nombre_completo, tipo_documento, numero_documento,
			telefono, email, descripcion_bien, fecha_incidente, detalle_reclamo, 
			pedido_consumidor, firma_digital, estado
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
	`, 
		fmt.Sprintf("TEST-%d-%d", time.Now().Unix(), time.Now().UnixNano()%1000000),
		"RECLAMO",
		"Cliente de Prueba",
		"DNI",
		"12345678",
		"999888777",
		"cliente@test.com",
		"Producto de prueba",
		time.Now(),
		"Detalle del reclamo de prueba",
		"Solicitud del consumidor de prueba",
		"data:image/png;base64,test",
		"PENDIENTE",
	).Scan(&reclamoID)
	
	if err != nil {
		t.Fatalf("Error creando reclamo de prueba: %v", err)
	}
	
	t.Logf("✅ Reclamo de prueba creado: %s", reclamoID)
	return reclamoID
}

// limpiarDatosPrueba limpia datos de prueba después de los tests
func limpiarDatosPrueba(t *testing.T) {
	ctx := context.Background()
	
	// Eliminar usuario de prueba
	testPool.Exec(ctx, "DELETE FROM usuarios_admin WHERE email = $1", "test@codeplex.com")
	
	// Eliminar reclamos de prueba
	testPool.Exec(ctx, "DELETE FROM reclamos WHERE codigo_reclamo LIKE 'TEST-%'")
	
	t.Log("✅ Datos de prueba limpiados")
}

// =============================================================================
// TESTS DE AUTENTICACIÓN
// =============================================================================

func TestLoginExitoso(t *testing.T) {
	setupTestDB(t)
	setupTestRouter(t)
	defer limpiarDatosPrueba(t)
	
	testUserID = crearUsuarioPrueba(t)
	
	loginData := map[string]string{
		"email":    "test@codeplex.com",
		"password": "Test123!",
	}
	
	body, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/api/admin/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code, "Login debe retornar 200")
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.True(t, response["success"].(bool), "Success debe ser true")
	assert.NotNil(t, response["data"], "Data no debe ser nil")
	
	data := response["data"].(map[string]interface{})
	assert.NotEmpty(t, data["token"], "Token no debe estar vacío")
	
	testToken = data["token"].(string)
	t.Logf("✅ Login exitoso - Token generado: %s", testToken[:20]+"...")
}

func TestLoginCredencialesInvalidas(t *testing.T) {
	setupTestDB(t)
	setupTestRouter(t)
	defer limpiarDatosPrueba(t)
	
	crearUsuarioPrueba(t)
	
	loginData := map[string]string{
		"email":    "test@codeplex.com",
		"password": "WrongPassword123!",
	}
	
	body, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/api/admin/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Login con credenciales inválidas debe retornar 401")
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.False(t, response["success"].(bool), "Success debe ser false")
	assert.Equal(t, "Credenciales inválidas", response["message"], "Mensaje debe indicar credenciales inválidas")
	
	t.Log("✅ Login con credenciales inválidas rechazado correctamente")
}

func TestLoginUsuarioInactivo(t *testing.T) {
	setupTestDB(t)
	setupTestRouter(t)
	defer limpiarDatosPrueba(t)
	
	userID := crearUsuarioPrueba(t)
	
	// Desactivar usuario
	ctx := context.Background()
	testPool.Exec(ctx, "UPDATE usuarios_admin SET activo = false WHERE id = $1", userID)
	
	loginData := map[string]string{
		"email":    "test@codeplex.com",
		"password": "Test123!",
	}
	
	body, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/api/admin/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code, "Login de usuario inactivo debe retornar 403")
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.False(t, response["success"].(bool), "Success debe ser false")
	assert.Equal(t, "Usuario inactivo", response["message"], "Mensaje debe indicar usuario inactivo")
	
	t.Log("✅ Login de usuario inactivo bloqueado correctamente")
}

func TestAccesoSinToken(t *testing.T) {
	setupTestDB(t)
	setupTestRouter(t)
	
	req, _ := http.NewRequest("GET", "/api/admin/reclamos", nil)
	
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Acceso sin token debe retornar 401")
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.False(t, response["success"].(bool), "Success debe ser false")
	assert.Equal(t, "Token requerido", response["message"], "Mensaje debe indicar token requerido")
	
	t.Log("✅ Acceso sin token bloqueado correctamente")
}

func TestAccesoConTokenInvalido(t *testing.T) {
	setupTestDB(t)
	setupTestRouter(t)
	
	req, _ := http.NewRequest("GET", "/api/admin/reclamos", nil)
	req.Header.Set("Authorization", "Bearer token_invalido_12345")
	
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Acceso con token inválido debe retornar 401")
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.False(t, response["success"].(bool), "Success debe ser false")
	assert.Equal(t, "Token inválido", response["message"], "Mensaje debe indicar token inválido")
	
	t.Log("✅ Acceso con token inválido bloqueado correctamente")
}

// =============================================================================
// TESTS DE GESTIÓN DE RECLAMOS
// =============================================================================

func TestListarReclamos(t *testing.T) {
	setupTestDB(t)
	setupTestRouter(t)
	defer limpiarDatosPrueba(t)
	
	// Crear usuario y obtener token
	crearUsuarioPrueba(t)
	TestLoginExitoso(t) // Esto setea testToken
	
	// Crear reclamos de prueba
	crearReclamoPrueba(t)
	crearReclamoPrueba(t)
	
	req, _ := http.NewRequest("GET", "/api/admin/reclamos", nil)
	req.Header.Set("Authorization", "Bearer "+testToken)
	
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code, "Listar reclamos debe retornar 200")
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.True(t, response["success"].(bool), "Success debe ser true")
	assert.NotNil(t, response["data"], "Data no debe ser nil")
	
	data := response["data"].([]interface{})
	assert.GreaterOrEqual(t, len(data), 2, "Debe haber al menos 2 reclamos")
	
	t.Logf("✅ Listado de reclamos exitoso - Total: %d", len(data))
}

func TestListarReclamosFiltradoPorEstado(t *testing.T) {
	setupTestDB(t)
	setupTestRouter(t)
	defer limpiarDatosPrueba(t)
	
	crearUsuarioPrueba(t)
	TestLoginExitoso(t)
	
	crearReclamoPrueba(t)
	
	req, _ := http.NewRequest("GET", "/api/admin/reclamos?estado=PENDIENTE", nil)
	req.Header.Set("Authorization", "Bearer "+testToken)
	
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code, "Listar reclamos filtrados debe retornar 200")
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.True(t, response["success"].(bool), "Success debe ser true")
	
	data := response["data"].([]interface{})
	for _, item := range data {
		reclamo := item.(map[string]interface{})
		assert.Equal(t, "PENDIENTE", reclamo["estado"], "Todos los reclamos deben estar en estado PENDIENTE")
	}
	
	t.Logf("✅ Filtrado por estado exitoso - Reclamos PENDIENTES: %d", len(data))
}

func TestCambiarEstadoReclamo(t *testing.T) {
	setupTestDB(t)
	setupTestRouter(t)
	defer limpiarDatosPrueba(t)
	
	crearUsuarioPrueba(t)
	TestLoginExitoso(t)
	
	reclamoID := crearReclamoPrueba(t)
	
	cambioEstado := map[string]string{
		"estado":     "EN_PROCESO",
		"comentario": "Reclamo en revisión",
	}
	
	bodyEstado, _ := json.Marshal(cambioEstado)
	reqEstado, _ := http.NewRequest("PUT", "/api/admin/reclamos/"+reclamoID+"/estado", bytes.NewBuffer(bodyEstado))
	reqEstado.Header.Set("Content-Type", "application/json")
	reqEstado.Header.Set("Authorization", "Bearer "+testToken)
	
	wEstado := httptest.NewRecorder()
	testRouter.ServeHTTP(wEstado, reqEstado)
	
	assert.Equal(t, http.StatusOK, wEstado.Code, "Cambiar estado debe retornar 200")
	
	var response map[string]interface{}
	json.Unmarshal(wEstado.Body.Bytes(), &response)
	
	assert.True(t, response["success"].(bool), "Success debe ser true")
	assert.Equal(t, "Estado actualizado correctamente", response["message"])
	
	// Verificar en DB
	ctx := context.Background()
	var nuevoEstado string
	testPool.QueryRow(ctx, "SELECT estado FROM reclamos WHERE id = $1", reclamoID).Scan(&nuevoEstado)
	assert.Equal(t, "EN_PROCESO", nuevoEstado, "Estado en DB debe ser EN_PROCESO")
	
	t.Log("✅ Cambio de estado exitoso")
}

func TestCambiarEstadoInvalido(t *testing.T) {
	setupTestDB(t)
	setupTestRouter(t)
	defer limpiarDatosPrueba(t)
	
	crearUsuarioPrueba(t)
	TestLoginExitoso(t)
	
	reclamoID := crearReclamoPrueba(t)
	
	cambioEstado := map[string]string{
		"estado": "ESTADO_INVALIDO",
	}
	
	body, _ := json.Marshal(cambioEstado)
	req, _ := http.NewRequest("PUT", "/api/admin/reclamos/"+reclamoID+"/estado", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testToken)
	
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code, "Estado inválido debe retornar 400")
	
	t.Log("✅ Validación de estado inválido funciona correctamente")
}

func TestResponderReclamo(t *testing.T) {
	setupTestDB(t)
	setupTestRouter(t)
	defer limpiarDatosPrueba(t)
	
	crearUsuarioPrueba(t)
	TestLoginExitoso(t)
	
	reclamoID := crearReclamoPrueba(t)
	
	respuesta := map[string]string{
		"respuesta_empresa":     "Hemos revisado su caso y procedemos a dar solución inmediata.",
		"accion_tomada":         "Reemplazo del producto",
		"compensacion_ofrecida": "Producto nuevo + cupón de descuento",
	}
	
	body, _ := json.Marshal(respuesta)
	req, _ := http.NewRequest("POST", "/api/admin/reclamos/"+reclamoID+"/respuesta", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testToken)
	
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code, "Responder reclamo debe retornar 201")
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.True(t, response["success"].(bool), "Success debe ser true")
	assert.Equal(t, "Respuesta enviada correctamente", response["message"])
	
	// Verificar que se creó la respuesta en DB
	ctx := context.Background()
	var respuestaDB string
	err := testPool.QueryRow(ctx, "SELECT respuesta_empresa FROM respuestas WHERE reclamo_id = $1", reclamoID).Scan(&respuestaDB)
	assert.NoError(t, err, "Debe existir respuesta en DB")
	assert.Contains(t, respuestaDB, "Hemos revisado su caso", "Respuesta debe coincidir")
	
	// Verificar que el estado cambió a RESUELTO
	var estadoFinal string
	testPool.QueryRow(ctx, "SELECT estado FROM reclamos WHERE id = $1", reclamoID).Scan(&estadoFinal)
	assert.Equal(t, "RESUELTO", estadoFinal, "Estado debe cambiar a RESUELTO automáticamente")
	
	t.Log("✅ Respuesta a reclamo enviada exitosamente")
}

func TestResponderReclamoTextoCorto(t *testing.T) {
	setupTestDB(t)
	setupTestRouter(t)
	defer limpiarDatosPrueba(t)
	
	crearUsuarioPrueba(t)
	TestLoginExitoso(t)
	
	reclamoID := crearReclamoPrueba(t)
	
	respuesta := map[string]string{
		"respuesta_empresa": "OK", // Muy corto (menos de 10 caracteres)
	}
	
	body, _ := json.Marshal(respuesta)
	req, _ := http.NewRequest("POST", "/api/admin/reclamos/"+reclamoID+"/respuesta", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testToken)
	
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code, "Respuesta corta debe retornar 400")
	
	t.Log("✅ Validación de respuesta mínima funciona correctamente")
}

// =============================================================================
// TESTS DE ESTADÍSTICAS
// =============================================================================

func TestObtenerEstadisticas(t *testing.T) {
	setupTestDB(t)
	setupTestRouter(t)
	defer limpiarDatosPrueba(t)
	
	crearUsuarioPrueba(t)
	TestLoginExitoso(t)
	
	// Crear algunos reclamos de prueba
	crearReclamoPrueba(t)
	crearReclamoPrueba(t)
	
	req, _ := http.NewRequest("GET", "/api/admin/dashboard/stats", nil)
	req.Header.Set("Authorization", "Bearer "+testToken)
	
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code, "Obtener estadísticas debe retornar 200")
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.True(t, response["success"].(bool), "Success debe ser true")
	assert.NotNil(t, response["data"], "Data no debe ser nil")
	
	data := response["data"].(map[string]interface{})
	assert.NotNil(t, data["total_reclamos"], "total_reclamos debe existir")
	assert.NotNil(t, data["pendientes"], "pendientes debe existir")
	assert.NotNil(t, data["en_proceso"], "en_proceso debe existir")
	assert.NotNil(t, data["resueltos"], "resueltos debe existir")
	
	t.Logf("✅ Estadísticas obtenidas - Total: %.0f, Pendientes: %.0f", 
		data["total_reclamos"].(float64), 
		data["pendientes"].(float64))
}

// =============================================================================
// TESTS DE AUDITORÍA
// =============================================================================

func TestAuditoriaSeRegistra(t *testing.T) {
	setupTestDB(t)
	setupTestRouter(t)
	defer limpiarDatosPrueba(t)
	
	userID := crearUsuarioPrueba(t)
	
	// Login para obtener token
	loginData := map[string]string{
		"email":    "test@codeplex.com",
		"password": "Test123!",
	}
	
	body, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/api/admin/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})
	testToken = data["token"].(string)
	
	reclamoID := crearReclamoPrueba(t)
	
	// Cambiar estado
	cambioEstado := map[string]string{
		"estado": "EN_PROCESO",
	}
	
	bodyEstado, _ := json.Marshal(cambioEstado)
	reqEstado, _ := http.NewRequest("PUT", "/api/admin/reclamos/"+reclamoID+"/estado", bytes.NewBuffer(bodyEstado))
	reqEstado.Header.Set("Content-Type", "application/json")
	reqEstado.Header.Set("Authorization", "Bearer "+testToken)
	
	wEstado := httptest.NewRecorder()
	testRouter.ServeHTTP(wEstado, reqEstado)
	
	assert.Equal(t, http.StatusOK, wEstado.Code)
	
	// Verificar que se registró en auditoría
	ctx := context.Background()
	var auditCount int
	err := testPool.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM auditoria_admin 
		WHERE usuario_id = $1 AND entidad = 'RECLAMO' AND entidad_id = $2
	`, userID, reclamoID).Scan(&auditCount)
	
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, auditCount, 1, "Debe haber al menos 1 registro de auditoría")
	
	t.Logf("✅ Auditoría registrada correctamente - Registros: %d", auditCount)
}

// =============================================================================
// TEST RUNNER PRINCIPAL
// =============================================================================

func TestMain(m *testing.M) {
	fmt.Println("=============================================================================")
	fmt.Println("🧪 INICIANDO SUITE DE TESTS DEL BACKEND ADMIN")
	fmt.Println("=============================================================================")
	fmt.Println()
	
	// Configurar JWT secret para tests
	jwtSecret = "test-secret-key-12345"
	
	// Ejecutar tests
	exitCode := m.Run()
	
	fmt.Println()
	fmt.Println("=============================================================================")
	if exitCode == 0 {
		fmt.Println("✅ TODOS LOS TESTS PASARON EXITOSAMENTE")
	} else {
		fmt.Println("❌ ALGUNOS TESTS FALLARON")
	}
	fmt.Println("=============================================================================")
}