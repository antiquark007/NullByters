# Wipe Tool

The Wipe Tool is a utility designed to securely erase data from devices, including Android phones and computer storage. This tool integrates with an Electron-based GUI application and a certificate management tool, providing a comprehensive solution for data wiping and verification.

## Features

- **Secure Wiping**: Supports multiple wiping methods including Zero Fill, Random Fill, and Shred + Zero.
- **Android Device Support**: Wipe user data from Android devices using ADB and Fastboot.
- **Computer Wiping**: Erase data from computer storage devices securely.
- **Verification**: Option to verify the wipe by checking the first sector of the device.
- **Certificate Generation**: Generate certificates for wipe operations, ensuring compliance and traceability.
- **User-Friendly GUI**: An Electron-based graphical interface for easy interaction with the wipe tool.

## Installation

1. Clone the repository:
   ```
   git clone <repository-url>
   cd DekstopApp
   ```

2. Install the required dependencies for the certificate tool:
   ```
   cd Cert_Tool
   pip install -r requirements.txt
   ```

3. Install the required Node.js packages for the Electron GUI:
   ```
   cd gui
   npm install
   ```

4. Ensure that the necessary tools for wiping (e.g., ADB, Fastboot) are installed on your system.

## Usage

### Wipe Tool

To use the wipe tool, run the following command:
```
python wipe.py
```
Follow the prompts to select the device and wiping method.

### Electron GUI

To start the Electron GUI, navigate to the `gui` directory and run:
```
npm start
```

### Certificate Tool

To use the certificate tool, run:
```
python main.py
```
This will allow you to manage certificates related to wipe operations.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.

## Contributing

Contributions are welcome! Please submit a pull request or open an issue for any enhancements or bug fixes.

## Acknowledgments

- Thanks to the contributors and the open-source community for their support and tools that made this project possible.