#!/usr/bin/env node
const { spawnSync } = require("node:child_process");
const path = require("node:path");
const fs = require("node:fs");

const binDir = path.join(__dirname);
const isWin = process.platform === "win32";
const binary = path.join(binDir, isWin ? "project-context.exe" : "project-context");

if (!fs.existsSync(binary)) {
  console.error("❌ project-context binary not found. Run `npm install` again.");
  process.exit(1);
}

const result = spawnSync(binary, process.argv.slice(2), {
  stdio: "inherit",
  env: { ...process.env },
});

process.exit(result.status ?? 0);
