import dotenv from 'dotenv';
import pkg from 'pg';
const { Pool } = pkg;

dotenv.config();

console.log('🔍 Probando conexión a PostgreSQL...\n');
console.log('DATABASE_URL:', process.env.DATABASE_URL);
console.log('');

const pool = new Pool({
    connectionString: process.env.DATABASE_URL,
    ssl: false
});

async function testConnection() {
    try {
        console.log('⏳ Intentando conectar...');
        const client = await pool.connect();
        console.log('✅ Conexión exitosa!\n');
        
        const result = await client.query('SELECT version()');
        console.log('📊 Versión de PostgreSQL:');
        console.log(result.rows[0].version);
        console.log('');
        
        const tables = await client.query(`
            SELECT table_name 
            FROM information_schema.tables 
            WHERE table_schema = 'public'
        `);
        console.log('📋 Tablas en la base de datos:');
        tables.rows.forEach(row => console.log('  -', row.table_name));
        
        client.release();
        pool.end();
        
        console.log('\n✅ TODO FUNCIONA CORRECTAMENTE');
        process.exit(0);
    } catch (error) {
        console.error('❌ ERROR:', error.message);
        console.error('\n🔍 Detalles del error:', error);
        pool.end();
        process.exit(1);
    }
}

testConnection();