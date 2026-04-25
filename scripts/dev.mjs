#!/usr/bin/env node

import { spawn } from 'node:child_process';
import { createInterface } from 'node:readline';
import { existsSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptPath = fileURLToPath(import.meta.url);
const repoRoot = path.resolve(path.dirname(scriptPath), '..');
const serverDir = path.join(repoRoot, 'server');
const adminDir = path.join(repoRoot, 'admin');
const webDir = path.join(repoRoot, 'web');

const children = new Set();
let shuttingDown = false;

validateArgs();
validateWorkspace();

await installFrontendDeps('admin', adminDir);
if (existsSync(path.join(webDir, 'package.json'))) {
  await installFrontendDeps('web', webDir);
}

startService('api', serverDir, 'air', ['-c', '.air.toml']);
startService('worker', serverDir, 'air', ['-c', '.air.worker.toml']);
startService('admin', adminDir, 'bun', ['dev']);

if (existsSync(path.join(webDir, 'package.json'))) {
  startService('web', webDir, 'bun', ['dev']);
} else {
  console.log('[dev] 未检测到 web/package.json，跳过 web 前端。');
}

console.log('[dev] 本地开发服务已启动。按 Ctrl+C 退出全部进程。');

process.on('SIGINT', () => shutdown(0));
process.on('SIGTERM', () => shutdown(0));
process.on('uncaughtException', (error) => {
  console.error(`[dev] 启动脚本异常：${error.message}`);
  shutdown(1);
});

function validateArgs() {
  const runtimeArgs = process.argv.slice(2);
  if (runtimeArgs.length > 0) {
    console.error(`[dev] 启动脚本不接受运行参数：${runtimeArgs.join(', ')}`);
    console.error('[dev] 请直接运行 node ./scripts/dev.mjs。');
    process.exit(1);
  }
}

function validateWorkspace() {
  const configPath = path.join(serverDir, 'config.yaml');
  const configExamplePath = path.join(serverDir, 'config.example.yaml');
  const adminPackagePath = path.join(adminDir, 'package.json');

  if (!existsSync(configPath)) {
    console.error('[dev] 缺少 server/config.yaml。');
    console.error('[dev] 请先从 server/config.example.yaml 复制一份，并按本机 MariaDB/JWT 配置修改。');
    console.error(`[dev] 示例配置路径：${configExamplePath}`);
    process.exit(1);
  }

  if (!existsSync(adminPackagePath)) {
    console.error('[dev] 缺少 admin/package.json，无法启动管理端。');
    process.exit(1);
  }
}

async function installFrontendDeps(name, cwd) {
  console.log(`[${name}] 检查并安装前端依赖...`);
  await runStep(name, cwd, 'bun', ['install']);
}

function runStep(name, cwd, command, commandArgs) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, commandArgs, {
      cwd,
      env: process.env,
      stdio: ['ignore', 'pipe', 'pipe'],
    });

    prefixStream(name, child.stdout);
    prefixStream(name, child.stderr);

    child.on('error', (error) => {
      reject(new Error(`${name} 执行失败：${error.message}`));
    });

    child.on('exit', (code) => {
      if (code === 0) {
        resolve();
        return;
      }
      reject(new Error(`${name} 命令退出，状态码：${code}`));
    });
  }).catch((error) => {
    console.error(`[dev] ${error.message}`);
    process.exit(1);
  });
}

function startService(name, cwd, command, commandArgs) {
  const child = spawn(command, commandArgs, {
    cwd,
    env: process.env,
    stdio: ['ignore', 'pipe', 'pipe'],
  });

  children.add(child);
  prefixStream(name, child.stdout);
  prefixStream(name, child.stderr);

  child.on('error', (error) => {
    console.error(`[${name}] 启动失败：${error.message}`);
    shutdown(1);
  });

  child.on('exit', (code, signal) => {
    children.delete(child);
    if (shuttingDown) {
      return;
    }

    const reason = signal ? `信号 ${signal}` : `状态码 ${code}`;
    console.error(`[${name}] 进程已退出：${reason}`);
    shutdown(code === 0 ? 0 : 1);
  });
}

function prefixStream(name, stream) {
  const reader = createInterface({ input: stream });
  reader.on('line', (line) => {
    console.log(`[${name}] ${line}`);
  });
}

function shutdown(code) {
  if (shuttingDown) {
    return;
  }

  shuttingDown = true;
  console.log('[dev] 正在停止本地开发进程...');

  for (const child of children) {
    if (!child.killed) {
      child.kill();
    }
  }

  setTimeout(() => process.exit(code), 500);
}
