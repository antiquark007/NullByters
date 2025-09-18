import os, json, uuid
from pathlib import Path
from datetime import datetime
# from utils.config import PRIVATE_KEY_PEM
# from utils.qrgen import make_qr_png
# from utils.pdfgen import generate_certificate_pdf
# from utils.sign import load_private_key, sign_json_bytes
# from utils.payload_utils import canonical_json


def save_certificates(cert: dict, out_dir: str = "/var/log/NullBytes"):
    try:
        os.makedirs(out_dir, exist_ok=True)
    except PermissionError:
        out_dir = "/tmp/NullBytes"
        os.makedirs(out_dir, exist_ok=True)

    # --- Save JSON certificate ---
    json_path = Path(out_dir) / f"{cert['uuid']}_{os.path.basename(cert['device'])}.json"
    with open(json_path, "w") as f:
        json.dump(cert, f, indent=4)

    # # --- Sign + prepare PDF ---
    # try:
    #     private_key = load_private_key(PRIVATE_KEY_PEM)
    #     sig = sign_json_bytes(private_key, canonical_json(cert))
    #     payload_obj = {"cert": cert, "sig": sig}
    #
    #     qr_url = f"https://verify.nullbytes.org/?cert_id={cert['uuid']}"
    #     qr_png = Path(f"/tmp/{cert['uuid']}_qr.png")
    #     make_qr_png(qr_url, qr_png)
    #
    #     pdf_path = Path(out_dir) / f"{cert['uuid']}_{os.path.basename(cert['device'])}.pdf"
    #     generate_certificate_pdf(cert, qr_png, qr_url, pdf_path, payload_obj=payload_obj)
    # except Exception as e:
    #     print(f"[WARN] PDF generation failed: {e}")
    #     pdf_path = None

    return str(json_path)

    # return str(json_path), str(pdf_path) if pdf_path else None
