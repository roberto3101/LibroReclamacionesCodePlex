import { defineConfig } from 'astro/config';
import tailwind from '@astrojs/tailwind';
import react from '@astrojs/react';
// CAMBIO 1: Importamos vercel en lugar de node
import vercel from '@astrojs/vercel/serverless';

export default defineConfig({
  integrations: [tailwind(), react()],
  
  // Mantenemos output server
  output: 'server',
  
  // CAMBIO 2: Usamos el adaptador de Vercel
  adapter: vercel(),

  site: 'https://codeplex.pe'
});