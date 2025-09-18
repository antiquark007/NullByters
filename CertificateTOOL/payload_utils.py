import json, base64, zlib

def canonical_json(obj) -> bytes:
    # Deterministic JSON for signing
    return json.dumps(obj, separators=(",", ":"), sort_keys=True).encode()

def make_embedded_payload(cert_obj: dict, signature_b64: str) -> dict:
    return {"cert": cert_obj, "sig": signature_b64}

def encode_fragment_payload(payload_obj: dict) -> str:
    # Compress then base64url-encode to keep the URL short
    raw = json.dumps(payload_obj, separators=(",", ":"), sort_keys=True).encode()
    compressed = zlib.compress(raw, level=9)
    b64 = base64.urlsafe_b64encode(compressed).decode().rstrip("=")
    return b64

def decode_fragment_payload(b64_url: str) -> dict:
    # Only for testing/local parity with the verifier site
    pad = "=" * (-len(b64_url) % 4)
    compressed = base64.urlsafe_b64decode(b64_url + pad)
    raw = zlib.decompress(compressed)
    return json.loads(raw.decode())
