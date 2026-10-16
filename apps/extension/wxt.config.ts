import { defineConfig } from "wxt";
import tailwindcss from "@tailwindcss/vite";
import path from "path";

// See https://wxt.dev/api/config.html
export default defineConfig({
    manifest: {
        version: "0.1",
        name: "Nook",
        description: "",
        permissions: ["storage"],
        host_permissions: [
            // "http://localhost/*",
            "*://starbattle.puzzlebaron.com/*"
        ],
    },
    webExt: {
        startUrls: ["https://starbattle.puzzlebaron.com"]
    },
    srcDir: "./src",
    modules: ["@wxt-dev/module-react"],
    vite: () => ({
        plugins: [tailwindcss()],
        resolve: {
            alias: {
                "@": path.resolve(__dirname, "./src"),
            },
        },
    }),
});
