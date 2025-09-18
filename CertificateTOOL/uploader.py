import base64, json, requests
from datetime import datetime
from config import GITHUB_TOKEN, GITHUB_USER, GITHUB_REPO, GITHUB_BRANCH, GITHUB_PAGES_BASE

print("[DEBUG] Entered upload_cert_data()")
def upload_cert_data(payload_obj: dict, filename: str | None = None) -> str | None:
    try:
        print("[DEBUG] Uploading cert data...")
        print("[DEBUG] Token:", GITHUB_TOKEN[:6], "...")  # Don't print full token
        if not GITHUB_TOKEN:
            return None
        if filename is None:
            ts = datetime.utcnow().strftime("%Y%m%d_%H%M%S")
            filename = f"cert_{ts}.json"

        api_url = f"https://api.github.com/repos/{GITHUB_USER}/{GITHUB_REPO}/contents/{filename}"
        content_b64 = base64.b64encode(json.dumps(payload_obj, indent=2).encode()).decode()
        body = {
            "message": f"Add {filename}",
            "content": content_b64,
            "branch": GITHUB_BRANCH
        }
        headers = {"Authorization": f"token {GITHUB_TOKEN}", "Accept": "application/vnd.github+json"}
        r = requests.put(api_url, headers=headers, json=body, timeout=20)
        if r.status_code in (200, 201):
            print("[DEBUG] Upload succeeded ✅")
            return f"{GITHUB_PAGES_BASE}/{filename}"
        else:
            print(f"[ERROR] GitHub upload failed ❌ Status: {r.status_code}")
            print(f"[ERROR] Response: {r.text}")
            return None
    except Exception as e:
        print(f"[ERROR] Exception during upload: {e}")
        return None
