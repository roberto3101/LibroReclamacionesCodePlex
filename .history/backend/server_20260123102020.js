pool.on('connect', () => {
    console.log('✅ Conectado a CockroachDB');
});

pool.on('error', (err) => {
    console.error('❌ Error en CockroachDB:', err);
});