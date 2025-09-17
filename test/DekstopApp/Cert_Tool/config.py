import os

# Configuration settings for the certificate tool
class Config:
    BASE_DIR = os.path.dirname(os.path.abspath(__file__))
    
    # Paths
    LOG_DIR = os.path.join(BASE_DIR, 'logs')
    CERT_DIR = os.path.join(BASE_DIR, 'certs')
    KEYS_DIR = os.path.join(BASE_DIR, 'keys')
    SAMPLE_JSON = os.path.join(BASE_DIR, 'sample.json')
    
    # Constants
    CERTIFICATE_EXPIRY_DAYS = 365
    QR_CODE_SIZE = 300
    PDF_OUTPUT_DIR = os.path.join(BASE_DIR, 'pdfs')
    
    # Ensure directories exist
    os.makedirs(LOG_DIR, exist_ok=True)
    os.makedirs(CERT_DIR, exist_ok=True)
    os.makedirs(PDF_OUTPUT_DIR, exist_ok=True)