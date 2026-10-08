import { fileURLToPath, URL } from "node:url";

import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import vueDevTools from "vite-plugin-vue-devtools";
import AutoImport from "unplugin-auto-import/vite";
import Components from "unplugin-vue-components/vite";
import { ElementPlusResolver } from "unplugin-vue-components/resolvers";
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
    // Element Plus 组件和样式均按需引入，避免所有页面加载完整组件库 CSS。
    AutoImport({
      resolvers: [ElementPlusResolver({ importStyle: "css" })],
    }),
    Components({
      resolvers: [ElementPlusResolver({ importStyle: "css" })],
    }),
    tailwindcss(),
  ],
  server: {
    port: 1001,
    strictPort: true,
    proxy: {
      '/api': {
        target: 'http://localhost:9090',
        changeOrigin: true,
      },
      '/uploads': {
        target: 'http://localhost:9090',
        changeOrigin: true,
      },
    },
  },
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
});
