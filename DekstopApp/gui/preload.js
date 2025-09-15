// preload.js — expose safe API to renderer
const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('api', {
  loadPage: (page) => ipcRenderer.invoke('load-page', page),
  scanDevices: () => ipcRenderer.invoke('scan-devices'),
  startWipe: (opts) => ipcRenderer.send('start-wipe', opts),
  onWipeProgress: (cb) => ipcRenderer.on('wipe-progress', (_e, d) => cb(d)),
  onWipeLog: (cb) => ipcRenderer.on('wipe-log', (_e, d) => cb(d)),
  onWipeDone: (cb) => ipcRenderer.on('wipe-done', (_e, d) => cb(d)),
  generateCert: (args) => ipcRenderer.invoke('generate-cert', args),
  verifyCert: (args) => ipcRenderer.invoke('verify-cert', args),
  showSaveDialog: (opts) => ipcRenderer.invoke('show-save-dialog', opts),
  copyFile: (args) => ipcRenderer.invoke('copy-file', args)
});
