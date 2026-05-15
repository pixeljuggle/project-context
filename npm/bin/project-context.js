#!/usr/bin/env node
const { spawnSync } = require("node:child_process");
const os = require("node:os");
const path = require("node:path");

const platform = os.platform();
const arch = os.arch();

// Map Node.js os/platform names → exact optional dependency package name
const knownPackages = {
  "darwin arm64": "@pixeljuggle/project-context-darwin-arm64",
  "darwin x64": "@pixeljuggle/project-context-darwin-amd64",
  "linux arm64": "@pixeljuggle/project-context-linux-arm64",
  "linux x64": "@pixeljuggle/project-context-linux-amd64",
  "win32 x64": "@pixeljuggle/project-context-windows-amd64",
};

const packageName = knownPackages[`${platform} ${arch}`];

if (!packageName) {
  console.error(`❌ Unsupported platform: ${platform} ${arch}`);
  process.exit(1);
}

try {
  // require.resolve gives us the exact location of the installed optional package
  const pkgPath = require.resolve(`${packageName}/package.json`);
  const binDir = path.dirname(pkgPath);

  const binaryName = platform === "win32" ? "project-context.exe" : "project-context";
  const binaryPath = path.join(binDir, binaryName);

  // Forward all arguments to the native binary
  const result = spawnSync(binaryPath, process.argv.slice(2), {
    stdio: "inherit",
    env: { ...process.env },
  });

  process.exit(result.status ?? 0);
} catch (error) {
  console.error(`❌ Failed to find or execute binary for ${packageName}.`);
  console.error(
    "This usually means the optional dependency was skipped (--no-optional) or failed to install.",
  );
  console.error("Try: npm install --include=optional");
  process.exit(1);
}
