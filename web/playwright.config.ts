import os from "node:os";
import path from "node:path";
import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  outputDir: path.join(os.tmpdir(), "ai-pixel-playwright-results"),
  fullyParallel: false,
  retries: 0,
  reporter: "line",
  use: {
    baseURL: "http://127.0.0.1:18080",
    trace: "retain-on-failure"
  },
  webServer: {
    command: "go run ../cmd/e2e-server -addr 127.0.0.1:18080",
    url: "http://127.0.0.1:18080/health/ready",
    timeout: 120_000,
    reuseExistingServer: false
  },
  projects: [
    { name: "desktop-chromium", use: { ...devices["Desktop Chrome"] } },
    { name: "mobile-chromium", use: { ...devices["Pixel 7"] } }
  ]
});
