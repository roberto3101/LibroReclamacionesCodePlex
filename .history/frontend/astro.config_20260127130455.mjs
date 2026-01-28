// astro.config.mjs
import { defineConfig } from 'astro/config';
import react from '@astrojs/react';
import tailwind from '@astrojs/tailwind';

// 1. IMPORTANTE: Importar el adaptador de Node
import node from '@astrojs/node';

export default defineConfig({
  // 2. IMPORTANTE: Cambiar output a 'server'
  output: 'server',

  // 3. IMPORTANTE: Configurar el adaptador
  adapter: node({
    mode: 'standalone',
  }),

  integrations: [react(), tailwind()],
});