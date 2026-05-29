import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// Build natijasi ../static papkaga chiqadi — Go backend o'shani tarqatadi.
// Dev rejimida /api so'rovlari local Go serverga (8080) yo'naltiriladi.
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: "../static",
    emptyOutDir: true,
  },
  server: {
    port: 5173,
    proxy: {
      "/api": "http://localhost:8080",
    },
  },
});
