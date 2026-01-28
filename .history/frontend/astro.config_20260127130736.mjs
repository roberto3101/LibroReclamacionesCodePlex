import { defineConfig } from 'astro/config';
import tailwind from '@astrojs/tailwind';
import react from '@astrojs/react';
// 1. Importamos el adaptador que acabas de instalar
import node from '@astrojs/node';

export default defineConfig({
  // 2. Cambiamos de 'static' a 'server' para que acepte los IDs de los reclamos en vivo
  output: 'server',
  
  // 3. Configuramos el adaptador para que Astro sepa cómo correr el servidor
  adapter: node({
    mode: 'standalone',
  }),

  integrations: [tailwind(), react()],
  site: 'https://codeplex.pe'
});