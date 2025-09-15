const back = document.getElementById('back');
const save = document.getElementById('save');
const reset = document.getElementById('reset');
const optHpa = document.getElementById('opt-hpa');
const optVerify = document.getElementById('opt-verify');
const optDry = document.getElementById('opt-dry');

back.addEventListener('click', () => window.api.loadPage('landing.html'));

function loadOpts() {
  const s = JSON.parse(sessionStorage.getItem('advancedOptions') || '{}');
  optHpa.checked = !!s.hpa;
  optVerify.checked = !!s.verify;
  optDry.checked = !!s.dry;
}
loadOpts();

save.addEventListener('click', () => {
  const opts = { hpa: optHpa.checked, verify: optVerify.checked, dry: optDry.checked };
  sessionStorage.setItem('advancedOptions', JSON.stringify(opts));
  alert('Advanced options saved.');
  window.api.loadPage('landing.html');
});

reset.addEventListener('click', () => {
  sessionStorage.removeItem('advancedOptions');
  loadOpts();
});
