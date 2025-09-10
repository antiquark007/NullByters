# My C Project

## Overview
This project is designed to provide a comprehensive solution for securely wiping data from devices. It includes functionalities for device scanning, safety checks, command execution, and logging operations in JSON format.

## Project Structure
```
my-c-project
├── src
│   ├── main.c            # Entry point of the application
│   ├── device_scan.c     # Device scanning logic
│   ├── device_scan.h     # Header for device_scan.c
│   ├── safety.c          # Safety checks and validations
│   ├── safety.h          # Header for safety.c
│   ├── exec_cmd.c        # Command execution logic
│   ├── exec_cmd.h        # Header for exec_cmd.c
│   ├── wipe_ops.c        # Core wipe operations logic
│   ├── wipe_ops.h        # Header for wipe_ops.c
│   ├── hpa_dco.c         # HPA and DCO management
│   ├── hpa_dco.h         # Header for hpa_dco.c
│   ├── json_log.c        # JSON logging functionality
│   ├── json_log.h        # Header for json_log.c
│   ├── nist_map.c        # NIST mapping operations
│   ├── nist_map.h        # Header for nist_map.c
│   └── util.h            # Utility functions and macros
├── include                # Additional shared header files
├── Makefile               # Build instructions
└── README.md              # Project documentation
```

## Setup Instructions
1. Clone the repository:
   ```
   git clone <repository-url>
   cd my-c-project
   ```

2. Build the project using the Makefile:
   ```
   make
   ```

3. Run the application:
   ```
   ./your_executable_name
   ```

## Usage
- The application can be used to scan devices and perform secure wipe operations.
- Refer to the individual source files for specific functionalities and usage instructions.

## Contributing
Contributions are welcome! Please feel free to submit a pull request or open an issue for any suggestions or improvements.

## License
This project is licensed under the MIT License. See the LICENSE file for details.