import fs from 'fs';
import pkg from 'pg';
const { Client } = pkg;

const sql = fs.readFileSync('schema.sql', 'utf8');
const client = new Client({
    connectionString: 'postgresql://postgres:sql@127.0.0.1:5432/libro_reclamaciones'
});

client.connect()
    .then(() => client.query(sql))
    .then(() => {
        console.log('✅ Schema ejecutado correctamente');
        return client.end();
    })
    .catch(err => {
        console.error('❌ Error:', err.message);
        client.end();
    });