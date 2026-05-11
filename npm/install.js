#!/usr/bin/env node

const https = require("node:https");
const fs = require("node:fs");
const path = require("node:path");
const zlib = require("node:zlib");
const tar = require("tar");
const { pipeline } = require("node:stream/promises");

const pkg = require("./package.json");
const version = pkg.version;
const binaryName = "project-context";
const repo = "pixeljuggle/project-context";

async function main() {
  const platform = process.platform;
  const arch = process.arch;

  const goOs = platform === "win32" ? "windows" : platform;
  let goArch = arch === "x64" ? "amd64" : arch;
  if (goArch === "arm") goArch = "arm64";

  // ← THIS WAS THE BUG: now matches your actual release filenames (hyphen)
  const archiveName = `${binaryName}_${version}_${goOs}-${goArch}.tar.gz`;
  const url = `https://github.com/${repo}/releases/download/v${version}/${archiveName}`;

  const binDir = path.join(__dirname, "bin");
  fs.mkdirSync(binDir, { recursive: true });

  const targetBinary = platform === "win32" ? `${binaryName}.exe` : binaryName;
  const targetPath = path.join(binDir, targetBinary);

  console.log(`📥 Downloading ${binaryName} ${version} for ${goOs}-${goArch}...`);

  const response = await new Promise((resolve, reject) => {
    https
      .get(url, (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          return https.get(res.headers.location, resolve).on("error", reject);
        }
        resolve(res);
      })
      .on("error", reject);
  });

  if (response.statusCode !== 200) {
    throw new Error(`Download failed with status ${response.statusCode}\nURL: ${url}`);
  }

  await pipeline(
    response,
    zlib.createGunzip(),
    tar.extract({
      cwd: binDir,
      filter: (header) => path.basename(header) === targetBinary,
      strip: 0,
    }),
  );

  if (platform !== "win32") {
    fs.chmodSync(targetPath, "755");
  }

  console.log(`${binaryName} ${version} installed successfully!`);
}

main()
  .then(() => process.exit(0))
  .catch((err) => {
    console.error("❌ Failed to install project-context:", err.message);
    process.exit(1);
  });
