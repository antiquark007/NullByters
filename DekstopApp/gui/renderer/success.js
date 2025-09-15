const summaryArea = document.getElementById('summaryArea');
const gen = document.getElementById('genCert');
const viewJson = document.getElementById('viewJson');
const exportBtn = document.getElementById('exportBtn');
const home = document.getElementById('home');
const certStatus = document.getElementById('certStatus');

const logPath = sessionStorage.getItem('lastWipeLog');
summaryArea.innerHTML = `<div class="small-muted">Wipe Log: ${logPath || 'N/A'}</div>`;

let lastCertPath = null;

gen.addEventListener('click', async () => {
  gen.disabled = true;
  certStatus.innerText = 'Generating certificate...';
  const outJson = `cert_${Date.now()}.json`;
  const res = await window.api.generateCert({ logPath, outJson, outPdf: null });
  if (res.ok) {
    lastCertPath = res.json;
    certStatus.innerText = `Certificate created: ${res.json}`;
    viewJson.disabled = false;
    exportBtn.disabled = false;
  } else {
    certStatus.innerText = `Failed: ${res.error || JSON.stringify(res)}`;
  }
  gen.disabled = false;
});

viewJson.addEventListener('click', async () => {
  if (!lastCertPath) return alert('No cert found');
  const saveRes = await window.api.showSaveDialog({ title: 'Save certificate JSON', defaultPath: lastCertPath });
  if (!saveRes.canceled) {
    // ask main to copy file from lastCertPath to chosen path
    const res = await window.api.copyFile({ src: lastCertPath, dst: saveRes.filePath });
    if (res.ok) alert('Saved to: ' + res.dst);
    else alert('Save failed: ' + res.error);
  }
});

exportBtn.addEventListener('click', async () => {
  if (!lastCertPath) return alert('No cert to export');
  const saveRes = await window.api.showSaveDialog({ title:'Export certificate', defaultPath: `certificate_${Date.now()}.json` });
  if (!saveRes.canceled) {
    const res = await window.api.copyFile({ src: lastCertPath, dst: saveRes.filePath });
    if (res.ok) alert('Exported to: ' + res.dst);
    else alert('Export failed: ' + res.error);
  }
});

home.addEventListener('click', () => window.api.loadPage('landing.html'));
