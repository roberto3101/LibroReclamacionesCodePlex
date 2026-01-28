package main

import (
    "context"
    "database/sql"
    "log"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
)

// GET /api/admin/reclamos/:id
func obtenerReclamoAdminHandler(c *gin.Context) {
    id := c.Param("id")
    log.Printf("🔍 ADMIN: Buscando reclamo por ID UUID: %s", id) // LOG DE DEBUG

    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()

    query := `
        SELECT 
            r.id, r.codigo_reclamo, r.tipo_solicitud, r.estado,
            r.nombre_completo, r.tipo_documento, r.numero_documento,
            r.telefono, r.email, r.domicilio, r.departamento, r.provincia, r.distrito,
            r.tipo_bien, r.monto_reclamado, r.descripcion_bien,
            r.area_queja, r.descripcion_situacion,
            r.fecha_incidente, r.detalle_reclamo, r.pedido_consumidor,
            r.fecha_registro, r.fecha_limite_respuesta,
            res.respuesta_empresa, res.respondido_por
        FROM reclamos r
        LEFT JOIN respuestas res ON r.id = res.reclamo_id
        WHERE r.id = $1
    `

    var r struct {
        ID, Codigo, Tipo, Estado, Nombre, TipoDoc, NumDoc, Tel, Email string
        Dom, Dep, Prov, Dist, TipoBien                           sql.NullString
        Monto                                                    float64
        DescBien, AreaQueja, DescSit                             sql.NullString
        FechaInc                                                 time.Time
        Detalle, Pedido                                          string
        FechaReg, FechaLim                                       time.Time
        Respuesta, RespondidoPor                                 sql.NullString
    }

    err := pool.QueryRow(ctx, query, id).Scan(
        &r.ID, &r.Codigo, &r.Tipo, &r.Estado, &r.Nombre, &r.TipoDoc, &r.NumDoc, &r.Tel, &r.Email,
        &r.Dom, &r.Dep, &r.Prov, &r.Dist, &r.TipoBien, &r.Monto, &r.DescBien,
        &r.AreaQueja, &r.DescSit, &r.FechaInc, &r.Detalle, &r.Pedido,
        &r.FechaReg, &r.FechaLim, &r.Respuesta, &r.RespondidoPor,
    )

    if err != nil {
        log.Printf("❌ ADMIN: Error buscando ID %s: %v", id, err) // LOG DE ERROR
        c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Reclamo no encontrado"})
        return
    }

    log.Printf("✅ ADMIN: Reclamo %s encontrado. Enviando datos.", r.Codigo) // LOG DE EXITO

    // Mapeo manual
    data := gin.H{
        "id":                     r.ID,
        "codigo_reclamo":         r.Codigo,
        "tipo_solicitud":         r.Tipo,
        "estado":                 r.Estado,
        "nombre_completo":        r.Nombre,
        "tipo_documento":         r.TipoDoc,
        "numero_documento":       r.NumDoc,
        "telefono":               r.Tel,
        "email":                  r.Email,
        "domicilio":              nullToInterface(r.Dom),
        "departamento":           nullToInterface(r.Dep),
        "provincia":              nullToInterface(r.Prov),
        "distrito":               nullToInterface(r.Dist),
        "tipo_bien":              nullToInterface(r.TipoBien),
        "monto_reclamado":        r.Monto,
        "descripcion_bien":       nullToInterface(r.DescBien),
        "area_queja":             nullToInterface(r.AreaQueja),
        "descripcion_situacion":  nullToInterface(r.DescSit),
        "detalle_reclamo":        r.Detalle,
        "pedido_consumidor":      r.Pedido,
        "fecha_incidente":        r.FechaInc,
        "fecha_registro":         r.FechaReg,
        "fecha_limite_respuesta": r.FechaLim,
        "respuesta_empresa":      nullToInterface(r.Respuesta),
        "respondido_por":         nullToInterface(r.RespondidoPor),
    }

    c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}