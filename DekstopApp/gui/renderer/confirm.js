const deviceSummary = document.getElementById('deviceSummary');
const back = document.getElementById('back');
const cancel = document.getElementById('cancel');
const startBtn = document.getElementById('start');
const confirmInput = document.getElementById('confirmInput');
const methodSelect = document.getElementById('methodSelect');

const selected = JSON.parse(sessionStorage.getItem('selectedDevice') || '{}');
deviceSummary.innerHTML = `<div class="hl">${selected.name || selected.path || 'Unknown device'}</div>
  <div class="small-muted">Path: ${selected.path || selected.devicePath || 'N/A'} • Serial: ${selected.serial || 'N/A'}</div>`;

back.addEventListener('click', ()=> window.api.loadPage('detect.html'));
cancel.addEventListener('click', ()=> window.api.loadPage('detect.html'));

confirmInput.addEventListener('input', () => {
  startBtn.disabled = confirmInput.value !== 'WIPE';
});

startBtn.addEventListener('click', () => {
  const p = selected.path || selected.devicePath || selected.path;
  if (!p || p === '/' || p.toLowerCase().startsWith('c:')) {
    alert('Refusing to wipe system drive. Choose a removable target.');
    return;
  }
  const method = methodSelect.value;
  const outLog = `wipe_log_${Date.now()}.json`;
  sessionStorage.setItem('wipeLogPath', outLog);
  sessionStorage.setItem('wipeMethod', method);
  window.api.startWipe({ devicePath: p, method, outputLog: outLog });
  window.api.loadPage('progress.html');
});
