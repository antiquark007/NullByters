const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('api', {
    wipeDevice: (device, method) => ipcRenderer.invoke('wipe-device', device, method),
    verifyWipe: (device) => ipcRenderer.invoke('verify-wipe', device),
    getDevices: () => ipcRenderer.invoke('list-devices'),
    showError: (message) => ipcRenderer.invoke('show-error', message),
    showInfo: (message) => ipcRenderer.invoke('show-info', message),
    showConfirmation: (message) => ipcRenderer.invoke('show-confirmation', message),
});