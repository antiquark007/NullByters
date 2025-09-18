from pathlib import Path
import os
GITHUB_TOKEN = os.getenv("GITHUB_TOKEN", "")

OUT_DIR = Path("out")
OUT_DIR.mkdir(exist_ok=True)

PUBLIC_KEY_PEM ="keys/public.pem"
PRIVATE_KEY_PEM ="keys/private.pem"

# GitHub auto-upload settings (for hosted link path)
# Set these via environment variables for safety.
GITHUB_TOKEN = os.getenv("GITHUB_TOKEN", "")
GITHUB_USER = os.getenv("GITHUB_USER", "AdityaRajj23")  # GitHub username
GITHUB_REPO = os.getenv("GITHUB_REPO", "Cert_data") # repo name
GITHUB_BRANCH = os.getenv("GITHUB_BRANCH", "main")
GITHUB_PAGES_BASE = f"https://{GITHUB_USER}.github.io/{GITHUB_REPO}"

# Verifier site base (GitHub Pages)
VERIFIER_BASE = f"https://{GITHUB_USER}.github.io/Verifier_site"

# QR code output size
QR_PX = 512
