import requests
import json
import os

def upload_certificate(cert_path, server_url):
    with open(cert_path, 'rb') as cert_file:
        files = {'file': cert_file}
        response = requests.post(server_url, files=files)
        return response.status_code, response.json()

def upload_log(log_path, server_url):
    with open(log_path, 'rb') as log_file:
        files = {'file': log_file}
        response = requests.post(server_url, files=files)
        return response.status_code, response.json()

def main():
    server_url = os.getenv('CERT_UPLOAD_URL')  # URL to upload certificates
    if not server_url:
        print("Error: CERT_UPLOAD_URL environment variable not set.")
        return

    # Example paths, these should be replaced with actual paths
    cert_path = 'path/to/certificate.json'
    log_path = 'path/to/wipe_log.json'

    # Upload certificate
    cert_status, cert_response = upload_certificate(cert_path, server_url)
    if cert_status == 200:
        print("Certificate uploaded successfully:", cert_response)
    else:
        print("Failed to upload certificate:", cert_response)

    # Upload log
    log_status, log_response = upload_log(log_path, server_url)
    if log_status == 200:
        print("Log uploaded successfully:", log_response)
    else:
        print("Failed to upload log:", log_response)

if __name__ == "__main__":
    main()