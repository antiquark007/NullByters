# Integration of Electron GUI and Wipe Tool

This directory contains the integration files that facilitate communication between the Electron GUI application and the Python wipe tool. The integration is designed to allow users to initiate wipe operations from the GUI and receive feedback on the process.

## Files Overview

- **bridge.js**: This file acts as a bridge between the Electron GUI and the Python wipe script. It handles the communication and data transfer between the two components, ensuring that commands from the GUI are executed in the Python environment.

- **python-ipc.js**: This file manages inter-process communication (IPC) between the Electron application and the Python script. It allows for sending commands to the Python script and receiving responses, enabling a seamless user experience.

## Usage Instructions

1. **Setup**: Ensure that the Electron application and the Python wipe tool are properly set up in their respective directories.

2. **Running the Application**: Start the Electron application. The GUI will provide options to initiate wipe operations.

3. **Wipe Operations**: When a user selects a wipe operation from the GUI, the command is sent to the Python script via the IPC mechanism. The Python script will execute the wipe operation and return the status to the GUI.

4. **Feedback**: The GUI will display progress and completion messages based on the responses received from the Python script.

## Requirements

- Node.js and npm for the Electron application.
- Python 3.x for the wipe tool.
- Necessary Python packages as specified in the `requirements.txt` file in the Cert_Tool directory.

## Conclusion

This integration allows for a user-friendly interface to perform device wipe operations securely and efficiently. For further details on the individual components, refer to their respective README files in the `Cert_Tool`, `gui`, and `wipe-tool` directories.