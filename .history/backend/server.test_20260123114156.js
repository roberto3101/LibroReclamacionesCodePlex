// PRUEBAS UNITARIAS - Libro de Reclamaciones
// Ejecutar: npm test

import request from 'supertest';

const API_URL = 'http://localhost:3000';
let codigoTest = '';

describe('API Libro de Reclamaciones', () => {
    
    // Test 1: Health Check
    test('Health check debe retornar connected', async () => {
        const res = await request(API_URL).get('/api/health');
        expect(res.status).toBe(200);
        expect(res.body.database).toBe('connected');
    });
    
    // Test 2: Crear reclamo válido
    test('Debe crear un reclamo válido', async () => {
        const reclamo = {
            tipo_solicitud: 'RECLAMO',
            nombre_completo: 'Test User',
            tipo_documento: 'DNI',
            numero_documento: '12345678',
            telefono: '987654321',
            email: 'test@test.com',
            tipo_bien: 'PRODUCTO',
            descripcion_bien: 'Test producto',
            fecha_incidente: '2026-01-20',
            detalle_reclamo: 'Test detalle',
            pedido_consumidor: 'Test pedido',
            firma_digital: 'data:image/png;base64,iVBORw0KGgoAAAANS',
            acepta_terminos: true
        };
        
        const res = await request(API_URL).post('/api/reclamos').send(reclamo);
        expect(res.status).toBe(201);
        expect(res.body.success).toBe(true);
        codigoTest = res.body.data.codigo_reclamo;
    });
    
    // Test 3: Rechazar sin firma
    test('Debe rechazar reclamo sin firma', async () => {
        const reclamo = {
            tipo_solicitud: 'RECLAMO',
            nombre_completo: 'Test',
            tipo_documento: 'DNI',
            numero_documento: '12345678',
            telefono: '987654321',
            email: 'test@test.com',
            tipo_bien: 'PRODUCTO',
            descripcion_bien: 'Test',
            fecha_incidente: '2026-01-20',
            detalle_reclamo: 'Test',
            pedido_consumidor: 'Test',
            acepta_terminos: true
        };
        
        const res = await request(API_URL).post('/api/reclamos').send(reclamo);
        expect(res.status).toBe(400);
    });
    
    // Test 4: Consultar reclamo
    test('Debe consultar reclamo existente', async () => {
        const res = await request(API_URL).get(`/api/reclamos/${codigoTest}`);
        expect(res.status).toBe(200);
        expect(res.body.data.codigo_reclamo).toBe(codigoTest);
    });
    
    // Test 5: Ver firma
    test('Debe ver firma como imagen', async () => {
        const res = await request(API_URL).get(`/api/reclamos/${codigoTest}/firma`);
        expect(res.status).toBe(200);
        expect(res.headers['content-type']).toBe('image/png');
    });




// Test 6: Verificar RUC y Razón Social Oficiales
    // Esto confirma que el DEFAULT de la base de datos funciona
    test('El reclamo guardado debe tener los datos fiscales de CODEPLEX', async () => {
        const res = await request(API_URL).get(`/api/reclamos/${codigoTest}`);
        
        expect(res.body.data.ruc).toBe('20539782232');
        expect(res.body.data.razon_social).toBe('CODEPLEX SOFTWARE S.A.C.');
    });









});