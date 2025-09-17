import os
import json
import subprocess
from cryptography.hazmat.backends import default_backend
from cryptography.hazmat.primitives import serialization, hashes
from cryptography.hazmat.primitives.asymmetric import padding

def load_private_key(key_path):
    with open(key_path, "rb") as key_file:
        private_key = serialization.load_pem_private_key(
            key_file.read(),
            password=None,
            backend=default_backend()
        )
    return private_key

def sign_data(data, private_key):
    signature = private_key.sign(
        data,
        padding.PSS(
            mgf=padding.MGF1(hashes.SHA256()),
            salt_length=padding.PSS.MAX_LENGTH
        ),
        hashes.SHA256()
    )
    return signature

def save_signature(signature, output_path):
    with open(output_path, "wb") as f:
        f.write(signature)

def sign_certificate(cert_data, private_key_path, output_path):
    private_key = load_private_key(private_key_path)
    signature = sign_data(cert_data.encode(), private_key)
    save_signature(signature, output_path)

def main():
    cert_data_path = "path/to/certificate.json"  # Update with actual path
    private_key_path = "keys/private.pem"
    output_signature_path = "path/to/signature.sig"  # Update with actual path

    with open(cert_data_path, "r") as f:
        cert_data = json.load(f)

    sign_certificate(json.dumps(cert_data), private_key_path, output_signature_path)

if __name__ == "__main__":
    main()