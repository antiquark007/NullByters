const deviceListEl = document.getElementById('deviceList');
const rescanBtn = document.getElementById('rescan');
const backHome = document.getElementById('backHome');

async function renderDevices() {
  deviceListEl.innerHTML = '<div class="small-muted">Scanning...</div>';
  try {
    const res = await window.api.scanDevices();
    deviceListEl.innerHTML = '';
    const devices = res.devices || [];
    if (devices.length === 0) {
      deviceListEl.innerHTML = '<div class="small-muted">No removable devices detected.</div>';
      return;
    }
    devices.forEach(d => {
      const li = document.createElement('li');
      li.className = 'device-item';
      li.innerHTML = `
        <div>
          <div class="hl">${d.name || d.device || d.path}</div>
          <div class="small">${d.size_gb || d.size || 'N/A'} GB • ${d.serial || 'unknown'}</div>
        </div>
        <div style="display:flex;gap:10px;align-items:center">
          <div class="small-muted">${d.type||'disk'}</div>
          <button class="btn ghost selectBtn">Select</button>
        </div>
      `;
      li.querySelector('.selectBtn').addEventListener('click', () => {
        sessionStorage.setItem('selectedDevice', JSON.stringify(d));
        window.api.loadPage('confirm.html');
      });
      deviceListEl.appendChild(li);
    });
  } catch (e) {
    deviceListEl.innerHTML = `<div class="warning">Scan failed: ${String(e)}</div>`;
  }
}

rescanBtn.addEventListener('click', renderDevices);
backHome.addEventListener('click', ()=> window.api.loadPage('landing.html'));

// initial
renderDevices();
