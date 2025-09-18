import argparse
import json
from pathlib import Path
from PyPDF2 import PdfReader
from sign import load_public_key, verify_json_bytes
from payload_utils import canonical_json
from config import PUBLIC_KEY_PEM


def extract_payload_from_pdf_metadata(pdf_path: Path) -> dict | None:
    """
    Reads the /CertPayload metadata field from the PDF (if present)
    and returns it as a dict.
    """
    try:
        reader = PdfReader(str(pdf_path))
        meta = reader.metadata or {}
        if "/CertPayload" in meta:
            return json.loads(meta["/CertPayload"])
    except Exception as e:
        print(f"[ERROR] Could not read payload from PDF metadata: {e}")
    return None


def main():
    parser = argparse.ArgumentParser(description="Offline PDF verifier for Secure Wipe Certificates")
    parser.add_argument("--pdf", required=True, help="Path to certificate PDF")
    args = parser.parse_args()

    pdf_path = Path(args.pdf)
    if not pdf_path.exists():
        print(f"[ERROR] PDF not found: {pdf_path}")
        return

    payload_obj = extract_payload_from_pdf_metadata(pdf_path)
    if not payload_obj:
        print("[ERROR] No embedded payload found in PDF metadata.")
        return

    if "cert" not in payload_obj or "sig" not in payload_obj:
        print("[ERROR] Invalid payload format in PDF metadata.")
        return

    public_key = load_public_key(PUBLIC_KEY_PEM)
    ok = verify_json_bytes(public_key, canonical_json(payload_obj["cert"]), payload_obj["sig"])

    if ok:
        print("✅ Certificate is VALID")
        print(json.dumps(payload_obj["cert"], indent=2))
    else:
        print("❌ Certificate verification FAILED")


if __name__ == "__main__":
    main()
