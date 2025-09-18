# main_generate.py
import json
import argparse
import socket
from pathlib import Path

from config import OUT_DIR, PRIVATE_KEY_PEM, VERIFIER_BASE
from sign import load_private_key, sign_json_bytes
from payload_utils import canonical_json, make_embedded_payload, encode_fragment_payload
from uploader import upload_cert_data
from qr_utils import make_qr_png
from pdf_gen import generate_certificate_pdf



def check_internet(host="github.com", port=443, timeout=3):
    """Quick check for internet connectivity."""
    try:
        socket.create_connection((host, port), timeout=timeout)
        print("[DEBUG] Internet check passed ✅")
        return True
    except OSError as e:
        print(f"[DEBUG] Internet check failed ❌: {e}")
        return False



def main():
    parser = argparse.ArgumentParser(description="Generate Secure Wipe Certificate with dual‑mode QR")
    parser.add_argument("--json", required=True, help="Path to wipe tool JSON file")
    parser.add_argument("--pdf-out", default=str(OUT_DIR / "certificate.pdf"), help="Output PDF path")
    parser.add_argument("--qr-out", default=str(OUT_DIR / "certificate.qr.png"), help="Output QR PNG path")
    parser.add_argument("--subtitle", default="Issued by NullBytes", help="Optional subtitle under title")
    parser.add_argument("--no-upload", action="store_true", help="Disable GitHub upload even if online")
    args = parser.parse_args()

    cert_json_path = Path(args.json)
    pdf_out = Path(args.pdf_out)
    qr_png_out = Path(args.qr_out)

    # Load cert data
    cert_obj = json.loads(cert_json_path.read_text(encoding="utf-8"))

    # Sign canonical JSON
    priv = load_private_key(PRIVATE_KEY_PEM)
    to_sign = canonical_json(cert_obj)
    signature_b64 = sign_json_bytes(priv, to_sign)

    # Build payload object
    payload_obj = make_embedded_payload(cert_obj, signature_b64)

    # Decide QR payload
    hosted_url = None
    if not args.no_upload and check_internet():
        hosted_url = upload_cert_data(payload_obj)

    if hosted_url:
        # Online payload: hosted link
        fragment = hosted_url
        print(f"[INFO] Uploaded to GitHub: {hosted_url}")
    else:
        # Offline payload: compressed+base64url JSON+sig
        fragment = encode_fragment_payload(payload_obj)
        if not hosted_url:
            print("[INFO] Using offline payload in QR (no internet or upload disabled)")

    # Construct dual‑mode QR target URL
    qr_url = f"{VERIFIER_BASE}/#{fragment}"

    # Generate QR image
    make_qr_png(qr_url, qr_png_out)

    # Generate PDF
    generate_certificate_pdf(cert_obj, qr_png_out,qr_url, pdf_out, subtitle=args.subtitle,payload_obj=payload_obj)

    # Save convenience artifacts
    OUT_DIR.joinpath("certificate.json").write_text(json.dumps(cert_obj, indent=2), encoding="utf-8")
    OUT_DIR.joinpath("certificate.sig.b64").write_text(signature_b64, encoding="utf-8")
    OUT_DIR.joinpath("certificate_qr_url.txt").write_text(qr_url, encoding="utf-8")

    print(f"[DONE] PDF: {pdf_out}")
    print(f"[DONE] QR: {qr_png_out}")
    print(f"[DONE] QR URL: {qr_url}")


if __name__ == "__main__":
    main()
