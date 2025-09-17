def verify_certificate(cert_path):
    import json
    import os

    if not os.path.exists(cert_path):
        return False, "Certificate file does not exist."

    with open(cert_path, 'r') as f:
        cert_data = json.load(f)

    # Example verification logic (this should be replaced with actual verification)
    if 'uuid' in cert_data and 'device' in cert_data:
        return True, "Certificate is valid."
    else:
        return False, "Certificate is invalid."

def main():
    import sys

    if len(sys.argv) != 2:
        print("Usage: python verifier.py <certificate_path>")
        sys.exit(1)

    cert_path = sys.argv[1]
    is_valid, message = verify_certificate(cert_path)

    if is_valid:
        print(f"Verification successful: {message}")
    else:
        print(f"Verification failed: {message}")

if __name__ == "__main__":
    main()