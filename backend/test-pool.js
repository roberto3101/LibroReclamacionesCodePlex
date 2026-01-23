import dotenv from 'dotenv';
import pkg from 'pg';
const { Pool } = pkg;

dotenv.config();

console.log('🔍 Probando Pool (igual que server.js)...\n');
console.log('DATABASE_URL:', process.env.DATABASE_URL);
console.log('NODE_ENV:', process.env.NODE_ENV);
console.log('');

const pool = new Pool({
    connectionString: process.env.DATABASE_URL,
    ssl: process.env.NODE_ENV === 'production' 
        ? { rejectUnauthorized: false }
        : false
});

console.log('📊 Estado del Pool:');
console.log('  totalCount:', pool.totalCount);
console.log('  idleCount:', pool.idleCount);
console.log('  waitingCount:', pool.waitingCount);
console.log('');

pool.on('connect', () => {
    console.log('✅ Evento "connect" disparado');
});

pool.on('error', (err) => {
    console.error('❌ Evento "error":', err.message);
});

async function testHealth() {
    console.log('⏳ Simulando endpoint /api/health...\n');
    
    // Esto es lo que hace el endpoint
    const status = {
        status: 'ok',
        timestamp: new Date().toISOString(),
        database: pool.totalCount > 0 ? 'connected' : 'disconnected'
    };
    
    console.log('📍 Respuesta del health check:');
    console.log(JSON.stringify(status, null, 2));
    console.log('');
    
    // Ahora intentemos hacer una query real
    console.log('⏳ Intentando query real...');
    try {
        const result = await pool.query('SELECT NOW()');
        console.log('✅ Query exitosa:', result.rows[0]);
        console.log('');
        console.log('📊 Estado del Pool DESPUÉS de query:');
        console.log('  totalCount:', pool.totalCount);
        console.log('  idleCount:', pool.idleCount);
    } catch (error) {
        console.error('❌ Error en query:', error.message);
    }
    
    await pool.end();
}

testHealth();