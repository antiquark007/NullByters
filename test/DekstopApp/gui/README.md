# Electron GUI Application for Device Wiping and Certificate Management

This project integrates a device wiping tool with an Electron GUI application and a certificate management tool. It allows users to securely wipe devices and manage certificates through a user-friendly interface.

## Project Structure

- **Cert_Tool/**: Contains the certificate management tool.
  - **config.py**: Configuration settings for the certificate tool.
  - **LICENSE**: Licensing information for the project.
  - **main.py**: Main entry point for the certificate tool application.
  - **payload_utils.py**: Utility functions for handling payloads.
  - **pdf_gen.py**: Functions for generating PDF documents.
  - **qr_utils.py**: Functions for generating and handling QR codes.
  - **README.md**: Documentation for the certificate tool.
  - **requirements.txt**: Python dependencies for the certificate tool.
  - **sample.json**: Sample JSON configuration or data file.
  - **sign.py**: Functions for signing certificates or data.
  - **uploader.py**: Functions for uploading certificates or related data.
  - **verifier.py**: Functions for verifying certificates or data integrity.
  - **wipe_log_1758042608481.json**: Log of wipe operations.
  - **.env**: Environment variables for configuration.
  - **.gitignore**: Files and directories to ignore by Git.
  - **cert_env/**: Virtual environment for the certificate tool.
  - **keys/**: Cryptographic keys used for signing and verification.

- **gui/**: Contains the Electron GUI application.
  - **index.js**: Main entry point for the Electron application.
  - **package.json**: Configuration file for the Electron application.
  - **preload.js**: Preload scripts for secure communication.
  - **README.md**: Documentation for the Electron GUI application.
  - **.gitignore**: Files and directories to ignore by Git.
  - **guide/**: Additional resources or guides.
  - **pages/**: HTML files for different pages of the application.
  - **renderer/**: JavaScript files for handling page logic.

- **wipe-tool/**: Contains the wipe tool.
  - **Makefile**: Commands for building and managing the wipe tool.
  - **README.md**: Documentation for the wipe tool.
  - **wipe-tool**: Compiled binary for the wipe tool.
  - **wipe-tool.c**: Source code for the wipe tool written in C.
  - **wipe.py**: Python script for wiping devices.

- **integration/**: Contains files for integrating the Electron GUI with the Python wipe script.
  - **bridge.js**: Bridge between the Electron GUI and the Python script.
  - **python-ipc.js**: Handles inter-process communication between the Electron application and the Python script.
  - **README.md**: Documentation for the integration.

## Getting Started

1. **Clone the Repository**: Clone this repository to your local machine.
2. **Install Dependencies**:
   - For the certificate tool, navigate to the `Cert_Tool` directory and run:
     ```
     pip install -r requirements.txt
     ```
   - For the Electron GUI, navigate to the `gui` directory and run:
     ```
     npm install
     ```
3. **Run the Applications**:
   - Start the certificate tool by running:
     ```
     python main.py
     ```
   - Start the Electron GUI application by running:
     ```
     npm start
     ```

## Usage

- Use the Electron GUI to select devices for wiping and manage certificates.
- Follow the prompts in the GUI to perform wipe operations and certificate management tasks.

## Contributing

Contributions are welcome! Please submit a pull request or open an issue for any enhancements or bug fixes.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.