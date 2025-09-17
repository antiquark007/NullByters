# Certificate Tool

The Certificate Tool is a Python application designed to manage and handle certificate operations, including signing, verifying, and generating related documents. This tool integrates with an Electron GUI for user-friendly interactions and also includes functionality for securely wiping devices.

## Features

- **Certificate Management**: Create, sign, and verify certificates.
- **PDF Generation**: Generate PDF documents for certificates and logs.
- **QR Code Handling**: Create and manage QR codes for certificate verification.
- **Device Wiping**: Securely wipe devices using integrated wipe functionality.
- **Logging**: Maintain logs of wipe operations for auditing and verification.

## Installation

1. Clone the repository:
   ```
   git clone <repository-url>
   cd Cert_Tool
   ```

2. Set up a virtual environment:
   ```
   cd cert_env
   source bin/activate
   ```

3. Install the required dependencies:
   ```
   pip install -r requirements.txt
   ```

4. Configure environment variables in the `.env` file as needed.

## Usage

- To run the certificate tool, execute:
  ```
  python main.py
  ```

- For wiping devices, use the integrated wipe functionality through the GUI or directly via the `wipe.py` script.

## Directory Structure

- `config.py`: Configuration settings for the application.
- `LICENSE`: Licensing information for the project.
- `main.py`: Main entry point for the application.
- `payload_utils.py`: Utility functions for handling payloads.
- `pdf_gen.py`: Functions for generating PDF documents.
- `qr_utils.py`: Functions for QR code generation and handling.
- `requirements.txt`: List of required Python packages.
- `sample.json`: Sample configuration or data file.
- `sign.py`: Functions for signing certificates.
- `uploader.py`: Functions for uploading certificates.
- `verifier.py`: Functions for verifying certificates.
- `wipe_log_*.json`: Logs of wipe operations.
- `cert_env/`: Virtual environment directory.
- `keys/`: Directory containing cryptographic keys.

## Contributing

Contributions are welcome! Please submit a pull request or open an issue for any enhancements or bug fixes.

## License

This project is licensed under the MIT License. See the LICENSE file for details.