// main.js - Electron main process
const { app, BrowserWindow, ipcMain, dialog } = require('electron');
const path = require('path');
const { spawn } = require('child_process');
const fs = require('fs');

let mainWindow;

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1100,
    height: 760,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      sandbox: false,
      nodeIntegration: false,
      contextIsolation: true
    }
  });

  mainWindow.loadFile(path.join(__dirname, 'pages', 'landing.html'));
}

app.whenReady().then(createWindow);
app.on('window-all-closed', () => { if (process.platform !== 'darwin') app.quit(); });
app.on('activate', () => { if (BrowserWindow.getAllWindows().length === 0) createWindow(); });

/* ---------- Helpers ---------- */

// stream child process stdout/stderr lines
function runCommandStream(cmd, args, callbacks = {}) {
  try {
    const p = spawn(cmd, args);
    p.stdout.setEncoding('utf8');
    p.stderr.setEncoding('utf8');
    p.stdout.on('data', (data) => {
      data.toString().split(/\r?\n/).forEach(line => { if (line) callbacks.onStdout?.(line); });
    });
    p.stderr.on('data', (data) => {
      data.toString().split(/\r?\n/).forEach(line => { if (line) callbacks.onStderr?.(line); });
    });
    p.on('close', (code) => callbacks.onExit?.(code));
    return p;
  } catch (err) {
    callbacks.onStderr?.(String(err));
    callbacks.onExit?.(-1);
    return null;
  }
}

/* ---------- IPC: page navigation ---------- */
ipcMain.handle('load-page', async (_ev, pageName) => {
  const file = path.join(__dirname, 'pages', pageName);
  if (!fs.existsSync(file)) {
    return { ok: false, error: 'Page not found' };
  }
  mainWindow.loadFile(file);
  return { ok: true };
});

/* ---------- IPC: device scan ---------- */
ipcMain.handle('scan-devices', async () => {
  // Path to your compiled wipe-tool; modify if needed
  const tool = path.join(__dirname, '..', 'wipe-tool', 'wipe-tool'); // ../wipe-tool/wipe-tool
  if (!fs.existsSync(tool)) {
    // mock mode for frontend development
    return {
      mock: true,
      devices: [
        { name: 'SanDisk USB 16GB', path: '/dev/sdb', size_gb: 16, serial: 'SN123' },
        { name: 'WD HDD 1TB', path: '/dev/sdc', size_gb: 1000, serial: 'WD001' }
      ]
    };
  }

  // real mode: call tool --list --json (tool should print JSON)
  return new Promise((resolve) => {
    const p = spawn(tool, ['--list', '--json']);
    let out = '', err = '';
    p.stdout.on('data', (d) => out += d);
    p.stderr.on('data', (d) => err += d);
    p.on('close', (code) => {
      if (code === 0) {
        try {
          const json = JSON.parse(out);
          resolve({ mock: false, devices: json.devices || [] });
        } catch (e) {
          resolve({ mock: false, devices: [], error: 'Invalid JSON from wipe-tool' });
        }
      } else {
        resolve({ mock: false, devices: [], error: err || `scan exited ${code}` });
      }
    });
  });
});

/* ---------- IPC: start wipe ---------- */
ipcMain.on('start-wipe', (ev, { devicePath, method, outputLog }) => {
  // Safety: block system drive by common path checks — extend this for your needs
  const lower = (devicePath || '').toString().toLowerCase();
  if (!devicePath || lower === '/' || lower.startsWith('c:') || lower.includes('\\\\') && lower.includes('c:')) {
    ev.reply('wipe-error', 'Refusing to wipe system/unsafe device');
    return;
  }

  const tool = path.join(__dirname, '..', 'wipe-tool', 'wipe-tool');
  if (!fs.existsSync(tool)) {
    // simulate progress in mock mode
    let percent = 0;
    const interval = setInterval(() => {
      percent += 10;
      ev.reply('wipe-progress', { percent, message: `Mock progress ${percent}%` });
      if (percent >= 100) {
        clearInterval(interval);
        const mockLog = {
          device: { path: devicePath, name: 'Mock USB', serial: 'MOCK' },
          wipe: { method, nist_level: method==='purge'?'purge':'clear', status:'success', started_at: new Date().toISOString(), finished_at: new Date().toISOString() },
          system: { tool_version: '0.1-demo' }
        };
        const outFile = outputLog || path.join(app.getPath('temp'), `mock_wipe_${Date.now()}.json`);
        fs.writeFileSync(outFile, JSON.stringify(mockLog, null, 2));
        ev.reply('wipe-done', { success: true, logPath: outFile });
      }
    }, 400);
    return;
  }

  // Real mode: spawn wipe-tool
  const args = ['--device', devicePath, '--method', method, '--output', outputLog || 'wipe_log.json'];
  runCommandStream(tool, args, {
    onStdout: (line) => {
      ev.reply('wipe-progress', { raw: line });
      // if line is JSON progress, try to parse and emit percent
      try {
        const o = JSON.parse(line);
        if (o.progress) ev.reply('wipe-progress', { percent: o.progress, message: o.message || '' });
      } catch (e) { /* ignore non-json */ }
    },
    onStderr: (line) => ev.reply('wipe-log', { level: 'error', text: line }),
    onExit: (code) => {
      if (code === 0) ev.reply('wipe-done', { success: true, logPath: outputLog || 'wipe_log.json' });
      else ev.reply('wipe-done', { success: false, code });
    }
  });
});

/* ---------- IPC: generate certificate ---------- */
ipcMain.handle('generate-cert', async (_ev, { logPath, outJson, outPdf }) => {
  const certGenBinary = path.join(__dirname, '..', 'certificates', 'cert_gen'); // compiled PyInstaller binary or script path
  if (!fs.existsSync(certGenBinary)) {
    // mock cert generation
    const mockCert = {
      device: { name: 'Mock USB', serial: 'MOCK' },
      wipe: { method: 'purge', status: 'success', started_at: new Date().toISOString(), finished_at: new Date().toISOString() },
      signature: { algorithm: 'Ed25519', sig: 'MOCKSIG' }
    };
    const jpath = outJson || path.join(app.getPath('temp'), `mock_cert_${Date.now()}.json`);
    fs.writeFileSync(jpath, JSON.stringify(mockCert, null, 2));
    return { ok: true, json: jpath, pdf: null, mock: true };
  }

  // Real invocation (if certGenBinary is a script/binary)
  return new Promise((resolve) => {
    const args = [logPath, '--out', outJson || 'cert.json', '--pdf', outPdf || 'cert.pdf'];
    const p = spawn(certGenBinary, args);
    let out = '', err = '';
    p.stdout.on('data', (d) => out += d);
    p.stderr.on('data', (d) => err += d);
    p.on('close', (code) => {
      if (code === 0) resolve({ ok: true, json: args[1], pdf: args[3] });
      else resolve({ ok: false, error: err || out });
    });
  });
});

/* ---------- IPC: verify certificate ---------- */
ipcMain.handle('verify-cert', async (_ev, { certPath, pubkey }) => {
  const verifier = path.join(__dirname, '..', 'certificates', 'verifier.py');
  if (!fs.existsSync(verifier)) {
    return { ok: true, verified: true, mock: true, output: 'Mock verified' };
  }
  return new Promise((resolve) => {
    const p = spawn('python3', [verifier, certPath, pubkey || '']);
    let out = '', err = '';
    p.stdout.on('data', (d) => out += d);
    p.stderr.on('data', (d) => err += d);
    p.on('close', (code) => {
      resolve({ ok: code === 0, output: out || err, code });
    });
  });
});

/* ---------- IPC: show save dialog ---------- */
ipcMain.handle('show-save-dialog', async (_ev, options) => {
  return dialog.showSaveDialog(mainWindow, options);
});

/* ---------- Optional: copy-file handler (renderer uses showSaveDialog then calls this) ---------- */
ipcMain.handle('copy-file', async (_ev, { src, dst }) => {
  try {
    fs.copyFileSync(src, dst);
    return { ok: true, dst };
  } catch (e) {
    return { ok: false, error: String(e) };
  }
});
