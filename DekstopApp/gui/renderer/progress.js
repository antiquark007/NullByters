const bar = document.getElementById('bar');
const percentLabel = document.getElementById('percentLabel');
const logBox = document.getElementById('log');
const etaLabel = document.getElementById('etaLabel');
const targetInfo = document.getElementById('targetInfo');
const homeBtn = document.getElementById('homeBtn');

const selected = JSON.parse(sessionStorage.getItem('selectedDevice') || '{}');
targetInfo.innerText = `${selected.name || selected.path || 'Unknown'} • ${selected.size_gb || selected.size || ''} GB`;

function appendLog(msg) {
  const d = document.createElement('div');
  d.textContent = `[${new Date().toLocaleTimeString()}] ${msg}`;
  logBox.appendChild(d);
  logBox.scrollTop = logBox.scrollHeight;
}

window.api.onWipeProgress((data) => {
  const percent = data.percent ?? (data.raw && (() => { try { const o = JSON.parse(data.raw); return o.progress; } catch(e){return null;} })());
  const message = data.message || data.raw || JSON.stringify(data);
  if (percent !== undefined && percent !== null) {
    bar.style.width = `${percent}%`;
    percentLabel.innerText = `${percent}%`;
  }
  appendLog(message);
});

window.api.onWipeDone((res) => {
  if (res.success) {
    appendLog('Wipe finished successfully.');
    sessionStorage.setItem('lastWipeLog', res.logPath || '');
    setTimeout(()=> window.api.loadPage('success.html'), 700);
  } else {
    appendLog('Wipe failed: ' + JSON.stringify(res));
    alert('Wipe failed: see logs.');
  }
});

homeBtn.addEventListener('click', ()=> window.api.loadPage('landing.html'));
appendLog('Waiting for progress messages...');
