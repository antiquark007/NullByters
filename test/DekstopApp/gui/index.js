const { app, BrowserWindow, ipcMain } = require('electron');
const path = require('path');
const { PythonShell } = require('python-shell');

let mainWindow;

function createWindow() {
    mainWindow = new BrowserWindow({
        width: 800,
        height: 600,
        webPreferences: {
            preload: path.join(__dirname, 'preload.js'),
            contextIsolation: true,
            enableRemoteModule: false,
            nodeIntegration: false,
        },
    });

    mainWindow.loadFile(path.join(__dirname, 'pages', 'landing.html'));

    mainWindow.on('closed', function () {
        mainWindow = null;
    });
}

app.on('ready', createWindow);

app.on('window-all-closed', function () {
    if (process.platform !== 'darwin') {
        app.quit();
    }
});

app.on('activate', function () {
    if (mainWindow === null) {
        createWindow();
    }
});

// IPC handlers for communication with Python script
ipcMain.on('start-wipe', (event, deviceType, wipeMethod) => {
    const options = {
        mode: 'text',
        pythonOptions: ['-u'],
        scriptPath: path.join(__dirname, '../wipe-tool'),
        args: [deviceType, wipeMethod],
    };

    PythonShell.run('wipe.py', options, (err, results) => {
        if (err) {
            event.reply('wipe-result', { success: false, message: err.message });
            return;
        }
        event.reply('wipe-result', { success: true, message: results });
    });
});