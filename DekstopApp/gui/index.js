const { app, BrowserWindow, ipcMain, dialog } = require('electron');
const path = require('path');
const { spawn } = require('child_process');
const fs = require('fs');

let mainWindow;

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1000,
    height: 700,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, 'preload.js')
    }
  });

  mainWindow.loadFile('pages/landing.html');
}

// Unified function to call Python script with proper error handling
function callPythonScript(args) {
    const scriptPath = path.join(__dirname, '..', 'wipe-tool', 'wipe.py');
    
    // Check if script exists
    if (!fs.existsSync(scriptPath)) {
        console.error('[ERROR] Python script not found at:', scriptPath);
        throw new Error(`Python script not found: ${scriptPath}`);
    }
    
    // Use correct Python command for platform
    const pythonCmd = process.platform === 'win32' ? 'python' : 'python3';
    
    console.log('[DEBUG] Executing:', pythonCmd, [scriptPath, ...args]);
    
    return spawn(pythonCmd, [scriptPath, ...args], {
        cwd: path.join(__dirname, '..'),
        env: process.env,
        stdio: ['pipe', 'pipe', 'pipe']
    });
}

/* ---------- IPC: Device Scanning ---------- */
ipcMain.handle('scan-devices', async () => {
  const pythonTool = path.join(__dirname, '..', 'wipe-tool', 'wipe.py');
  
  console.log('[DEBUG] Looking for Python tool at:', pythonTool);
  console.log('[DEBUG] File exists:', fs.existsSync(pythonTool));
  
  if (!fs.existsSync(pythonTool)) {
    console.log('[DEBUG] Python wipe tool not found, using mock data');
    return {
      devices: [
        {
          name: "Mock USB Drive",
          path: "\\\\?\\PHYSICALDRIVE1",  // Windows format for mock
          size: "16.0G",
          model: "Mock USB",
          serial: "MOCK123",
          removable: true,
          vendor: "Mock"
        }
      ],
      count: 1,
      timestamp: new Date().toISOString(),
      mock: true
    };
  }

  return new Promise((resolve) => {
    try {
      // Use the unified function
      const scanProcess = callPythonScript(['--list', '--json']);
      
      let stdout = '';
      let stderr = '';
      
      scanProcess.stdout.on('data', (data) => {
        stdout += data.toString();
      });
      
      scanProcess.stderr.on('data', (data) => {
        stderr += data.toString();
        console.error('[PYTHON ERROR]:', data.toString());
      });
      
      scanProcess.on('close', (code) => {
        console.log('[DEBUG] Scan process closed with code:', code);
        console.log('[DEBUG] Stdout:', stdout);
        console.log('[DEBUG] Stderr:', stderr);
        
        if (code === 0) {
          try {
            const result = JSON.parse(stdout);
            resolve(result);
          } catch (e) {
            console.error('[ERROR] Failed to parse JSON:', e);
            resolve({ devices: [], count: 0, error: 'Failed to parse scan results', details: e.message });
          }
        } else {
          resolve({ 
            devices: [], 
            count: 0, 
            error: stderr || `Scan failed with exit code ${code}`,
            code: code
          });
        }
      });
      
      scanProcess.on('error', (error) => {
        console.error('[ERROR] Scan process error:', error);
        resolve({ 
          devices: [], 
          count: 0, 
          error: `Process spawn error: ${error.message}`,
          code: 'SPAWN_ERROR'
        });
      });
      
    } catch (error) {
      console.error('[ERROR] Failed to start scan process:', error);
      resolve({ 
        devices: [], 
        count: 0, 
        error: error.message,
        code: 'SCRIPT_NOT_FOUND'
      });
    }
  });
});

/* ---------- IPC: Wipe Process ---------- */
ipcMain.on('start-wipe', (event, { devicePath, method, outputLog, deviceInfo }) => {
  const pythonTool = path.join(__dirname, '..', 'wipe-tool', 'wipe.py');
  const logPath = outputLog || path.join(app.getPath('temp'), `wipe_log_${Date.now()}.json`);
  
  // Safety checks
  const lower = (devicePath || '').toString().toLowerCase();
  if (!devicePath || lower === '/' || lower.startsWith('c:') || lower.includes('physicaldrive0')) {
    event.reply('wipe-error', 'Refusing to wipe system/unsafe device');
    return;
  }

  if (!fs.existsSync(pythonTool)) {
    // Mock mode
    let percent = 0;
    const interval = setInterval(() => {
      percent += 5;
      event.reply('wipe-progress', { 
        progress: percent, 
        message: `Mock wipe progress ${percent}%`
      });
      
      if (percent >= 100) {
        clearInterval(interval);
        
        const mockLog = {
          version: "1.0",
          device: { 
            path: devicePath, 
            name: deviceInfo?.name || 'Mock USB Device',
            size: deviceInfo?.size || '16.0G'
          },
          wipe: { 
            method, 
            status: 'success', 
            started_at: new Date().toISOString(), 
            finished_at: new Date().toISOString()
          }
        };
        
        fs.writeFileSync(logPath, JSON.stringify(mockLog, null, 2));
        
        event.reply('wipe-done', { 
          success: true, 
          logPath: logPath,
          devicePath: devicePath,
          deviceInfo: mockLog.device,
          method: method,
          mock: true
        });
      }
    }, 800);
    return;
  }

  // Real mode - use unified function
  try {
    const wipeProcess = callPythonScript(['--device', devicePath, '--method', method, '--output', logPath]);
    
    wipeProcess.stdout.on('data', (data) => {
      const lines = data.toString().split('\n');
      for (const line of lines) {
        if (line.trim()) {
          try {
            const progressData = JSON.parse(line);
            if (progressData.progress !== undefined) {
              event.reply('wipe-progress', progressData);
            }
          } catch (e) {
            event.reply('wipe-log', { level: 'info', text: line });
          }
        }
      }
    });
    
    wipeProcess.stderr.on('data', (data) => {
      console.error('[WIPE ERROR]:', data.toString());
      event.reply('wipe-log', { level: 'error', text: data.toString() });
    });
    
    wipeProcess.on('close', (code) => {
      if (code === 0) {
        let logData = {};
        try {
          if (fs.existsSync(logPath)) {
            logData = JSON.parse(fs.readFileSync(logPath, 'utf8'));
          }
        } catch (e) {
          console.error(`Failed to read wipe log: ${e.message}`);
        }
        
        event.reply('wipe-done', { 
          success: true, 
          logPath: logPath,
          devicePath: devicePath,
          deviceInfo: logData.device || deviceInfo,
          method: method
        });
      } else {
        event.reply('wipe-done', { 
          success: false, 
          error: `Wipe process failed with exit code ${code}`
        });
      }
    });
    
    wipeProcess.on('error', (error) => {
      console.error('[WIPE PROCESS ERROR]:', error);
      event.reply('wipe-done', { 
        success: false, 
        error: `Process error: ${error.message}`
      });
    });
    
  } catch (error) {
    console.error('[WIPE START ERROR]:', error);
    event.reply('wipe-done', { 
      success: false, 
      error: error.message
    });
  }
});

/* ---------- IPC: Certificate Generation ---------- */
ipcMain.handle('generate-cert', async (event, args) => {
  const { logPath, outJson, outPdf, deviceInfo } = args;
  
  const certToolDir = path.join(__dirname, '..', 'Cert_Tool');
  
  // Windows uses Scripts instead of bin
  const venvPython = process.platform === 'win32' 
    ? path.join(certToolDir, 'cert_env', 'Scripts', 'python.exe')
    : path.join(certToolDir, 'cert_env', 'bin', 'python');
    
  const mainScript = path.join(certToolDir, 'main.py');
  
  console.log('[DEBUG] Looking for venv python at:', venvPython);
  console.log('[DEBUG] Looking for main script at:', mainScript);
  
  if (!fs.existsSync(venvPython) || !fs.existsSync(mainScript)) {
    console.log('[DEBUG] Certificate tool not found, using mock');
    const mockCert = {
      version: "1.0",
      certificate: {
        id: `MOCK-${Date.now()}`,
        device: deviceInfo || { name: 'Mock USB' }
      }
    };
    
    const jpath = outJson || path.join(app.getPath('temp'), `mock_cert_${Date.now()}.json`);
    fs.writeFileSync(jpath, JSON.stringify(mockCert, null, 2));
    
    return { 
      success: true, 
      certificate_id: mockCert.certificate.id,
      jsonPath: jpath, 
      mock: true 
    };
  }

  const outputJson = outJson || path.join(app.getPath('temp'), `cert_${Date.now()}.json`);
  const outputPdf = outPdf || path.join(app.getPath('temp'), `cert_${Date.now()}.pdf`);
  
  const certArgs = [
    mainScript,
    '--input', logPath,
    '--output-json', outputJson,
    '--output-pdf', outputPdf,
    '--device-path', deviceInfo?.path || '/dev/unknown',
    '--device-name', deviceInfo?.name || 'Unknown Device',
    '--wipe-method', deviceInfo?.method || 'destroy'
  ];

  return new Promise((resolve) => {
    const certProcess = spawn(venvPython, certArgs, {
      stdio: ['pipe', 'pipe', 'pipe'],
      cwd: certToolDir
    });
    
    let stdout = '';
    let stderr = '';
    
    certProcess.stdout.on('data', (data) => {
      stdout += data.toString();
    });
    
    certProcess.stderr.on('data', (data) => {
      stderr += data.toString();
    });
    
    certProcess.on('close', (code) => {
      if (code === 0) {
        resolve({
          success: true,
          certificate_id: `CERT-${Date.now()}`,
          jsonPath: outputJson,
          pdfPath: outputPdf
        });
      } else {
        resolve({
          success: false,
          error: stderr || `Process exited with code ${code}`
        });
      }
    });
  });
});

/* ---------- Other IPC handlers ---------- */
ipcMain.handle('verify-cert', async (event, args) => {
  return { valid: true, message: 'Mock verification successful' };
});

ipcMain.handle('show-save-dialog', async (event, options) => {
  return await dialog.showSaveDialog(mainWindow, options);
});

ipcMain.handle('copy-file', async (event, { source, destination }) => {
  try {
    fs.copyFileSync(source, destination);
    return { success: true };
  } catch (error) {
    return { success: false, error: error.message };
  }
});

ipcMain.handle('load-page', async (event, pageName) => {
  const pagePath = path.join(__dirname, 'pages', pageName);
  if (fs.existsSync(pagePath)) {
    mainWindow.loadFile(pagePath);
    return { success: true };
  }
  return { success: false, error: 'Page not found' };
});

/* ---------- App Lifecycle ---------- */
app.whenReady().then(createWindow);

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});

app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0) {
    createWindow();
  }
});