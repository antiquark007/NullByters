const certFileEl = document.getElementById('certFile');
const verifyBtn = document.getElementById('verify');
const out = document.getElementById('verifyOutput');
const home = document.getElementById('home');

verifyBtn.addEventListener('click', async () => {
  out.innerText = 'Preparing...';
  const f = certFileEl.files[0];
  if (!f) { out.innerText = 'Select a certificate JSON file first'; return; }
  const certPath = f.path; // works in Electron file input
  const res = await window.api.verifyCert({ certPath, pubkey: null });
  out.innerText = JSON.stringify(res, null, 2);
});

home.addEventListener('click', ()=> window.api.loadPage('landing.html'));
