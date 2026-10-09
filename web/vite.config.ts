import preact from "@preact/preset-vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [preact(), tailwindcss()],
  server: {
    host: "127.0.0.1",
    port: 5173,
    strictPort: true,
  },
  build: {
    target: "es2022",
    assetsDir: "assets",
    manifest: true,
    sourcemap: false,
    modulePreload: { polyfill: false },
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes("@codemirror") || id.includes("/codemirror")) {
            return "editor";
          }
        },
      },
    },
  },
  test: {
    environment: "happy-dom",
  },
});
