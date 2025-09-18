from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import padding
from cryptography.hazmat.backends import default_backend
import base64

def load_private_key(pem_path):
    with open(pem_path, "rb") as f:
        return serialization.load_pem_private_key(f.read(), password=None, backend=default_backend())

def load_public_key(pem_path):
    with open(pem_path, "rb") as f:
        return serialization.load_pem_public_key(f.read(), backend=default_backend())

def sign_json_bytes(private_key, data_bytes: bytes) -> str:
    signature = private_key.sign(
        data_bytes,
        padding.PKCS1v15(),
        hashes.SHA256()
    )
    return base64.b64encode(signature).decode()

def verify_json_bytes(public_key, data_bytes: bytes, signature_b64: str) -> bool:
    try:
        public_key.verify(
            base64.b64decode(signature_b64),
            data_bytes,
            padding.PKCS1v15(),
            hashes.SHA256()
        )
        return True
    except Exception:
        return False
