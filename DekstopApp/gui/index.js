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
    
    // FOR LINUX: Use sudo to run Python script with elevated privileges
    if (process.platform !== 'win32') {
        return spawn('sudo', [pythonCmd, scriptPath, ...args], {
            cwd: path.join(__dirname, '..'),
            env: process.env,
            stdio: ['pipe', 'pipe', 'pipe']
        });
    } else {
        return spawn(pythonCmd, [scriptPath, ...args], {
            cwd: path.join(__dirname, '..'),
            env: process.env,
            stdio: ['pipe', 'pipe', 'pipe']
        });
    }
}


/* ---------- IPC: Device Scanning ---------- */
ipcMain.handle('scan-devices', async () => {
  // Use mainWipe.py as entry point for cross-platform support
  const pythonTool = path.join(__dirname, '..', 'wipe-tool', 'mainWipe.py');
  
  if (!fs.existsSync(pythonTool)) {
    console.log('[DEBUG] Main wipe tool not found, using mock data');
    return {
      devices: [
        {
          name: "Mock USB Drive",
          path: process.platform === 'win32' ? '\\\\.\\PHYSICALDRIVE1' : "/dev/sdb",
          size: "16.0G",
          model: "Mock USB",
          serial: "MOCK123",
          removable: true,
          vendor: "Mock",
          device_type: "USB Drive"
        }
      ],
      count: 1,
      timestamp: new Date().toISOString(),
      platform: process.platform,
      mock: true
    };
  }

  return new Promise((resolve) => {
    const scanProcess = spawn('python3', [pythonTool, '--list', '--json'], {
      stdio: ['pipe', 'pipe', 'pipe']
    });
    
    let stdout = '';
    let stderr = '';
    
    scanProcess.stdout.on('data', (data) => {
      stdout += data.toString();
    });
    
    scanProcess.stderr.on('data', (data) => {
      stderr += data.toString();
    });
    
    scanProcess.on('close', (code) => {
      console.log(`[DEBUG] Device scan completed with code: ${code}`);
      
      if (code === 0) {
        try {
          const result = JSON.parse(stdout);
          resolve(result);
        } catch (e) {
          console.error('[ERROR] Failed to parse device scan JSON:', e);
          resolve({ devices: [], count: 0, error: 'Failed to parse scan results' });
        }
      } else {
        console.error('[ERROR] Device scan failed:', stderr);
        resolve({ devices: [], count: 0, error: stderr || 'Scan failed' });
      }
    });
  });
});

/* ---------- IPC: Wipe Process ---------- */
ipcMain.on('start-wipe', (event, { devicePath, method, outputLog, deviceInfo }) => {
  const pythonTool = path.join(__dirname, '..', 'wipe-tool', 'mainWipe.py');
  const logPath = outputLog || path.join(app.getPath('temp'), `wipe_log_${Date.now()}.json`);
  
  // Platform-specific safety checks
  const isSystemPath = process.platform === 'win32' 
    ? devicePath.includes('PHYSICALDRIVE0')
    : (devicePath === '/' || devicePath.startsWith('/dev/sda'));
  
  if (!devicePath || isSystemPath) {
    event.reply('wipe-error', 'Refusing to wipe system/unsafe device');
    return;
  }

  if (!fs.existsSync(pythonTool)) {
    // Enhanced mock mode with platform detection
    console.log('[DEBUG] Main wipe tool not found, running enhanced mock wipe');
    let percent = 0;
    const platform = process.platform === 'win32' ? 'Windows' : 'Linux';
    
    const interval = setInterval(() => {
      percent += 3;
      event.reply('wipe-progress', { 
        progress: percent, 
        message: `Mock ${platform} wipe progress ${percent}%`,
        timestamp: new Date().toISOString(),
        platform: platform
      });
      
      if (percent >= 100) {
        clearInterval(interval);
        
        // Create enhanced mock log
        const mockLog = {
          version: "1.0",
          device: { 
            path: devicePath, 
            name: deviceInfo?.name || `Mock ${platform} Device`,
            size: deviceInfo?.size || '16.0G',
            serial: deviceInfo?.serial || 'MOCK123',
            model: deviceInfo?.model || `Mock ${platform} Drive`,
            vendor: deviceInfo?.vendor || 'Mock',
            device_type: deviceInfo?.device_type || 'Mock Drive'
          },
          wipe: { 
            method, 
            nist_level: method === 'purge' ? 'purge' : method === 'destroy' ? 'destroy' : 'clear',
            status: 'success', 
            started_at: new Date().toISOString(), 
            finished_at: new Date().toISOString(),
            passes_completed: method === 'destroy' ? 7 : method === 'purge' ? 3 : 1,
            verified_clean: true,
            tools_used: platform === 'Windows' ? ['PowerShell', 'FileStream API'] : ['dd', 'shred']
          },
          system: { 
            tool_version: `1.0.0-${platform.toLowerCase()}-mock`,
            platform: platform,
            operator: process.env.USERNAME || process.env.USER || 'Mock User',
            admin_privileges: true
          },
          compliance: {
            nist_800_88: true,
            certificate_id: `MOCK-${platform.toUpperCase()}-${Date.now()}`,
            dod_5220_22_m: method !== 'clear'
          }
        };
        
        fs.writeFileSync(logPath, JSON.stringify(mockLog, null, 2));
        
        event.reply('wipe-done', { 
          success: true, 
          logPath: logPath,
          devicePath: devicePath,
          deviceInfo: mockLog.device,
          method: method,
          timestamp: new Date().toISOString(),
          platform: platform,
          mock: true
        });
      }
    }, 600);
    return;
  }

  // Real mode with platform detection
  const args = [pythonTool, '--device', devicePath, '--method', method, '--output', logPath];
  
  console.log(`[DEBUG] Starting cross-platform wipe: python3 ${args.join(' ')}`);
  
  const wipeProcess = spawn('python3', args, {
    stdio: ['pipe', 'pipe', 'pipe']
  });
  
  wipeProcess.stdout.on('data', (data) => {
    const lines = data.toString().split('\n');
    for (const line of lines) {
      if (line.trim()) {
        try {
          const progressData = JSON.parse(line);
          if (progressData.progress !== undefined) {
            event.reply('wipe-progress', progressData);
          } else if (progressData.message) {
            event.reply('wipe-log', { level: 'info', text: progressData.message });
          }
        } catch (e) {
          // Non-JSON output, treat as log
          event.reply('wipe-log', { level: 'info', text: line });
        }
      }
    }
  });
  
  wipeProcess.stderr.on('data', (data) => {
    console.error(`[WIPE ERROR] ${data}`);
    event.reply('wipe-log', { level: 'error', text: data.toString() });
  });
  
  wipeProcess.on('close', (code) => {
    console.log(`[DEBUG] Cross-platform wipe process exited with code: ${code}`);
    
    if (code === 0) {
      // Read the generated log file
      let logData = {};
      try {
        if (fs.existsSync(logPath)) {
          logData = JSON.parse(fs.readFileSync(logPath, 'utf8'));
        }
      } catch (e) {
        console.error(`[ERROR] Failed to read wipe log: ${e.message}`);
      }
      
      event.reply('wipe-done', { 
        success: true, 
        logPath: logPath,
        devicePath: devicePath,
        deviceInfo: logData.device || deviceInfo,
        method: method,
        timestamp: new Date().toISOString(),
        platform: logData.system?.platform || 'Unknown',
        wipeData: logData
      });
    } else {
      event.reply('wipe-done', { 
        success: false, 
        code: code,
        error: `Cross-platform wipe process failed with exit code ${code}`
      });
    }
  });
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