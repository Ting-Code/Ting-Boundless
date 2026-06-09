import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

const gateway = process.env.VITE_GATEWAY_URL ?? 'http://127.0.0.1:8080';

const proxyCommon = {
  target: gateway,
  changeOrigin: true,
  cookieDomainRewrite: 'localhost',
};

export default defineConfig({
  plugins: [react()],
  base: '/admin/',
  server: {
    port: 5173,
    proxy: {
      '/v1': proxyCommon,
      '/sign-in': proxyCommon,
      '/callback': proxyCommon,
      '/sign-out': proxyCommon,
    },
  },
});
