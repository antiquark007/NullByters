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
    
    // Use correct Python command for platform - FIXED FOR WINDOWS
    const pythonCmd = process.platform === 'win32' ? 'python' : 'python3';
    
    console.log('[DEBUG] Executing:', pythonCmd, [scriptPath, ...args]);
    
    // FOR WINDOWS: Remove sudo and use direct python execution
    if (process.platform === 'win32') {
        return spawn(pythonCmd, [scriptPath, ...args], {
            cwd: path.join(__dirname, '..'),
            env: process.env,
            stdio: ['pipe', 'pipe', 'pipe'],
            shell: true  // Add shell: true for Windows
        });
    } else {
        // FOR LINUX: Use sudo to run Python script with elevated privileges
        return spawn('sudo', [pythonCmd, scriptPath, ...args], {
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
          interface: "USB",
          device_type: "USB Drive"
        }
      ],
      count: 1,
      mock: true
    };
  }

  return new Promise((resolve) => {
    try {
      // Use corrected Python command for Windows
      const pythonCmd = process.platform === 'win32' ? 'python' : 'python3';
      const scanProcess = spawn(pythonCmd, [pythonTool, '--list', '--json'], {
        cwd: path.join(__dirname, '..'),
        shell: process.platform === 'win32'  // Enable shell for Windows
      });

      let output = '';
      let stderr = '';

      scanProcess.stdout.on('data', (data) => {
        output += data.toString();
      });

      scanProcess.stderr.on('data', (data) => {
        stderr += data.toString();
        console.error('[SCAN ERROR]', data.toString());
      });

      scanProcess.on('close', (code) => {
        console.log(`[DEBUG] Scan process exited with code: ${code}`);
        if (code === 0 && output.trim()) {
          try {
            const result = JSON.parse(output);
            resolve(result);
          } catch (e) {
            console.error('[ERROR] Failed to parse scan results:', e);
            resolve({ devices: [], count: 0, error: 'Failed to parse scan results' });
          }
        } else {
          console.error('[ERROR] Device scan failed:', stderr);
          resolve({ devices: [], count: 0, error: stderr || 'Scan failed' });
        }
      });
    } catch (error) {
      console.error('[ERROR] Failed to start scan process:', error);
      resolve({ devices: [], count: 0, error: error.message });
    }
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
        
        try {
          fs.writeFileSync(logPath, JSON.stringify(mockLog, null, 2));
        } catch (err) {
          console.error('[ERROR] Failed to write mock log:', err);
        }
        
        event.reply('wipe-done', { 
          success: true, 
          logPath: logPath,
          devicePath: devicePath,
          deviceInfo: mockLog.device,
          method: method,
          timestamp: new Date().toISOString(),
          platform: platform,
          wipeData: mockLog,
          mock: true
        });
      }
    }, 600);
    return;
  }

  // Real mode with cross-platform support
  console.log(`[DEBUG] Starting ${process.platform === 'win32' ? 'Windows' : 'Linux'} wipe process`);
  
  // Use correct Python command for platform
  const pythonCmd = process.platform === 'win32' ? 'python' : 'python3';
  const args = [pythonTool, '--device', devicePath, '--method', method, '--output', logPath];
  
  console.log(`[DEBUG] Executing: ${pythonCmd} ${args.join(' ')}`);
  
  // Platform-specific spawn options
  const spawnOptions = {
    stdio: ['pipe', 'pipe', 'pipe'],
    cwd: path.join(__dirname, '..'),
    env: process.env
  };
  
  // Add Windows-specific options
  if (process.platform === 'win32') {
    spawnOptions.shell = true;
  }
  
  const wipeProcess = spawn(pythonCmd, args, spawnOptions);
  
  // Handle stdout data (progress updates and logs)
  wipeProcess.stdout.on('data', (data) => {
    const lines = data.toString().split('\n');
    for (const line of lines) {
      if (line.trim()) {
        try {
          const progressData = JSON.parse(line);
          
          // Handle different types of JSON messages
          if (progressData.progress !== undefined) {
            // Progress update
            event.reply('wipe-progress', progressData);
          } else if (progressData.message) {
            // Status message
            event.reply('wipe-log', { 
              level: 'info', 
              text: progressData.message,
              timestamp: progressData.timestamp || new Date().toISOString()
            });
          } else if (progressData.error) {
            // Error message
            event.reply('wipe-log', { 
              level: 'error', 
              text: progressData.error,
              timestamp: progressData.timestamp || new Date().toISOString()
            });
          }
        } catch (e) {
          // Non-JSON output, treat as regular log
          if (line.trim().length > 0) {
            event.reply('wipe-log', { 
              level: 'info', 
              text: line.trim(),
              timestamp: new Date().toISOString()
            });
          }
        }
      }
    }
  });
  
  // Handle stderr data (errors and debug info)
  wipeProcess.stderr.on('data', (data) => {
    const errorText = data.toString().trim();
    console.error(`[WIPE ERROR] ${errorText}`);
    
    // Try to parse as JSON first, fallback to plain text
    try {
      const errorData = JSON.parse(errorText);
      if (errorData.error) {
        event.reply('wipe-error', errorData.error);
      } else {
        event.reply('wipe-log', { 
          level: 'error', 
          text: errorText,
          timestamp: new Date().toISOString()
        });
      }
    } catch (e) {
      // Plain text error
      event.reply('wipe-log', { 
        level: 'error', 
        text: errorText,
        timestamp: new Date().toISOString()
      });
    }
  });
  
  // Handle process completion
  wipeProcess.on('close', (code) => {
    console.log(`[DEBUG] Cross-platform wipe process exited with code: ${code}`);
    
    if (code === 0) {
      // Success - read the generated log file
      let logData = {};
      try {
        if (fs.existsSync(logPath)) {
          const logContent = fs.readFileSync(logPath, 'utf8');
          logData = JSON.parse(logContent);
          console.log('[DEBUG] Successfully read wipe log file');
        } else {
          console.warn('[WARN] Wipe log file not found at:', logPath);
        }
      } catch (e) {
        console.error(`[ERROR] Failed to read wipe log: ${e.message}`);
      }
      
      event.reply('wipe-done', { 
        success: true, 
        logPath: logPath,
        devicePath: devicePath,
        deviceInfo: logData.device || deviceInfo || { name: 'Unknown Device' },
        method: method,
        timestamp: new Date().toISOString(),
        platform: logData.system?.platform || (process.platform === 'win32' ? 'Windows' : 'Linux'),
        wipeData: logData,
        exitCode: code
      });
    } else {
      // Failure
      const platform = process.platform === 'win32' ? 'Windows' : 'Linux';
      const errorMessage = `${platform} wipe process failed with exit code ${code}`;
      
      console.error(`[ERROR] ${errorMessage}`);
      
      event.reply('wipe-done', { 
        success: false, 
        code: code,
        error: errorMessage,
        devicePath: devicePath,
        method: method,
        timestamp: new Date().toISOString(),
        platform: platform
      });
    }
  });
  
  // Handle process errors (e.g., spawn failures)
  wipeProcess.on('error', (error) => {
    console.error(`[ERROR] Failed to start wipe process: ${error.message}`);
    
    // Provide helpful error messages based on the error type
    let userMessage = `Failed to start wipe process: ${error.message}`;
    
    if (error.code === 'ENOENT') {
      const platform = process.platform === 'win32' ? 'Windows' : 'Linux';
      const pythonCmd = process.platform === 'win32' ? 'python' : 'python3';
      userMessage = `Python not found. Please ensure ${pythonCmd} is installed and available in PATH on ${platform}.`;
    } else if (error.code === 'EACCES') {
      userMessage = 'Permission denied. Please run the application as Administrator (Windows) or with sudo (Linux).';
    }
    
    event.reply('wipe-error', userMessage);
  });
  
  // Send initial status
  event.reply('wipe-log', { 
    level: 'info', 
    text: `Starting wipe process for device: ${devicePath}`,
    timestamp: new Date().toISOString()
  });
  
  event.reply('wipe-log', { 
    level: 'info', 
    text: `Method: ${method.toUpperCase()} (${method === 'clear' ? '1 pass' : method === 'purge' ? '3 passes' : '7 passes'})`,
    timestamp: new Date().toISOString()
  });
  
  event.reply('wipe-log', { 
    level: 'info', 
    text: `Platform: ${process.platform === 'win32' ? 'Windows' : 'Linux'}`,
    timestamp: new Date().toISOString()
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