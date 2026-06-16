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
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules/react-dom') || id.includes('node_modules/react/')) {
            return 'vendor-react';
          }
          if (id.includes('node_modules/react-router')) {
            return 'vendor-router';
          }
          if (id.includes('node_modules/@tanstack/react-query')) {
            return 'vendor-query';
          }
          if (id.includes('node_modules/antd') || id.includes('node_modules/@ant-design')) {
            return 'vendor-antd';
          }
          if (id.includes('node_modules/rc-')) {
            return 'vendor-antd';
          }
        },
      },
    },
  },
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
