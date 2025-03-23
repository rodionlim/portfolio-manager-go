import { defineConfig } from "vite";
import { vanillaExtractPlugin } from "@vanilla-extract/vite-plugin";
import fs from "fs";
import react from "@vitejs/plugin-react";

// Read version from file
const version = fs.readFileSync("../../VERSION", "utf-8").trim();

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react(), vanillaExtractPlugin()],
  define: {
    "import.meta.env.APP_VERSION": JSON.stringify(version),
  },
  build: {
    outDir: "static",
  },
});
