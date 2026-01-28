import { defineConfig } from 'astro/config';
import tailwind from '@astrojs/tailwind';
import react from '@astrojs/react';
// 1. Importamos el adaptador
import node from '@astrojs/node';

export default defineConfig({
  integrations: [tailwind(), react()],
  
  // 2. Cambiamos a modo servidor (SSR) para soportar IDs dinámicos
  output: 'server',
  
  // 3. Configuramos el adaptador de Node
  adapter: node({
    mode: 'standalone',
  }),

  site: 'https://codeplex.pe'
});