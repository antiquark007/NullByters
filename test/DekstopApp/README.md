# Project Overview

This project is a comprehensive application designed to facilitate secure data wiping for both computers and Android devices, along with a certificate management tool. It consists of three main components: the Wipe Tool, the Certificate Tool, and an Electron-based GUI.

## Components

### 1. Wipe Tool
- **Location:** `wipe-tool/`
- **Description:** A Python script (`wipe.py`) that securely wipes data from devices. It integrates with the wipe tool written in C for enhanced performance.
- **Key Files:**
  - `wipe.py`: Main script for wiping operations.
  - `wipe-tool.c`: Source code for the C-based wipe tool.
  - `Makefile`: Build instructions for the wipe tool.

### 2. Certificate Tool
- **Location:** `Cert_Tool/`
- **Description:** A set of Python scripts for managing certificates, including signing, verifying, and generating QR codes and PDFs.
- **Key Files:**
  - `main.py`: Entry point for the certificate tool application.
  - `sign.py`: Functions for signing certificates.
  - `verifier.py`: Functions for verifying certificates.
  - `pdf_gen.py`: Functions for generating PDF documents.
  - `qr_utils.py`: Functions for handling QR codes.

### 3. Electron GUI
- **Location:** `gui/`
- **Description:** A user-friendly interface built with Electron to interact with the wipe tool and certificate tool.
- **Key Files:**
  - `index.js`: Main entry point for the Electron application.
  - `package.json`: Configuration file for the Electron app.
  - `renderer/`: Contains JavaScript files for handling page logic.
  - `pages/`: Contains HTML files for different views in the application.

### 4. Integration
- **Location:** `integration/`
- **Description:** Scripts that facilitate communication between the Electron GUI and the Python wipe script.
- **Key Files:**
  - `bridge.js`: Bridges the Electron GUI and the Python script.
  - `python-ipc.js`: Handles inter-process communication.

## Setup Instructions

1. **Clone the Repository:**
   ```
   git clone <repository-url>
   cd DekstopApp
   ```

2. **Set Up the Certificate Tool:**
   - Navigate to the `Cert_Tool` directory.
   - Install dependencies:
     ```
     pip install -r requirements.txt
     ```

3. **Set Up the Electron GUI:**
   - Navigate to the `gui` directory.
   - Install Node.js dependencies:
     ```
     npm install
     ```

4. **Run the Application:**
   - Start the Electron application:
     ```
     npm start
     ```

## Usage

- Use the GUI to select the type of device you want to wipe (computer or Android).
- Follow the prompts to securely wipe data.
- The Certificate Tool can be used to manage certificates, including signing and verification.

## License

This project is licensed under the MIT License. See the LICENSE file in each component for more details.