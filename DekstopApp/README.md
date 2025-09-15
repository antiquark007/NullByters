# NullBytes Desktop App
## Enterprise-Grade Secure Data Destruction Tool

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![NIST Compliant](https://img.shields.io/badge/NIST-800--88%20Compliant-green.svg)](https://csrc.nist.gov/publications/detail/sp/800-88/rev-1/final)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux-lightgrey.svg)](README.md)

---

## Overview

NullBytes is a comprehensive secure data destruction solution that combines military-grade wiping algorithms with an intuitive desktop interface. Built on a robust C backend with an Electron GUI frontend, it delivers NIST 800-88 compliant data sanitization for enterprise and individual use.

### Key Capabilities
- **NIST 800-88 Rev. 1** compliant sanitization methods
- **Multi-pass secure wiping** with cryptographic verification
- **Real-time progress monitoring** and compliance reporting
- **Cross-platform compatibility** (Windows/Linux)
- **Certificate generation** for audit trails

---

## Architecture

```
DesktopApp/
├── wipe-tool/              # C Backend - Secure Wiping Engine
│   ├── wipe-tool.c         # Core implementation
│   ├── Makefile            # Build configuration
│   ├── wipe-tool           # Compiled binary
│   └── README.md           # Backend documentation
└── gui/                    # Electron Frontend - User Interface
    ├── index.js            # Main process
    ├── preload.js          # Security bridge
    ├── package.json        # Dependencies
    ├── pages/              # UI templates
    ├── renderer/           # Frontend controllers
    └── guide/              # Documentation
```

---

## Installation

### System Requirements
- **OS**: Windows 10+ or Ubuntu 18.04+
- **RAM**: 512MB minimum
- **Storage**: 100MB free space
- **Permissions**: Administrator/root access required

### Quick Start

#### Ubuntu/Debian
```bash
# Install dependencies
sudo apt-get update
sudo apt-get install build-essential nodejs npm

# Optional: GUI support for headless systems
sudo apt-get install xvfb

# Clone and setup
cd /path/to/DesktopApp

# Build C backend
cd wipe-tool
make clean && make

# Install GUI dependencies
cd ../gui
npm install

# Launch application
npm start

# For root execution (add security flag)
npm start -- --no-sandbox
```

#### Headless Environment
```bash
# Run with virtual display
xvfb-run -a npm start
```

---

## Core Components

### Backend Engine (`wipe-tool/`)

The C-based backend provides high-performance secure wiping operations:

#### Key Functions
| Function | Purpose |
|----------|---------|
| `scan_devices()` | Enumerate removable storage devices |
| `wipe_device()` | Execute secure data destruction |
| `print_progress()` | Stream real-time progress updates |
| `create_wipe_log()` | Generate compliance audit logs |
| `is_system_drive()` | Prevent accidental system drive targeting |

#### Security Features
- **System Drive Protection**: Prevents targeting of boot/system partitions
- **Multi-Pattern Wiping**: Zeros, ones, and cryptographic random data
- **JSON Output**: Structured data for GUI integration
- **Cross-Platform**: Native device enumeration for Windows/Linux

### Frontend Interface (`gui/`)

Modern Electron-based GUI with security-focused architecture:

#### Main Process (`index.js`)
- Window management and lifecycle
- Secure IPC communication
- Backend process orchestration

#### Security Bridge (`preload.js`)
Exposes controlled API to renderer processes:
```javascript
window.api.scanDevices()    // Device enumeration
window.api.startWipe()      // Initiate wiping
window.api.onWipeProgress() // Progress monitoring
window.api.generateCert()   // Certificate creation
```

#### User Interface Pages
| Page | Purpose |
|------|---------|
| `landing.html` | Device selection and method configuration |
| `detect.html` | Device scanning interface |
| `confirm.html` | Safety confirmation dialog |
| `progress.html` | Real-time wiping progress |
| `success.html` | Completion summary and results |
| `verify.html` | Certificate verification |
| `advanced.html` | Expert configuration options |

---

## Security Standards

### NIST 800-88 Compliance Levels

| Level | Description | Passes | Pattern |
|-------|-------------|--------|---------|
| **Clear** | Basic sanitization | 1 | Zero fill (0x00) |
| **Purge** | Enhanced security | 3 | Multi-pattern overwrite |
| **Destroy** | Maximum security | 7 | Military-grade destruction |

### Secure Patterns
- `PATTERN_ZEROS`: Complete zero fill (0x00)
- `PATTERN_ONES`: Complete one fill (0xFF)
- `PATTERN_RANDOM`: Cryptographically secure random data

### System Protection
```c
int is_system_drive(const char *path) {
    // Validates against:
    // - Root filesystem mounts
    // - System partition detection  
    // - Removable device verification
}
```

---

## Usage Guide

### Command Line Interface
```bash
# List available devices
./wipe-tool --list --json

# Execute secure wipe with logging
./wipe-tool --device /dev/sdb --method destroy --output audit.json

# Quick clear operation
./wipe-tool --device /dev/sdc --method clear
```

### Graphical Interface Workflow

1. **Launch Application**
   ```bash
   npm start
   ```

2. **Device Detection**
   - Click "Scan Devices" 
   - Review detected removable storage

3. **Configuration**
   - Select target device
   - Choose sanitization method
   - Configure advanced options

4. **Safety Confirmation**
   - Type "WIPE" to confirm operation
   - Final verification prompt

5. **Execution & Monitoring**
   - Real-time progress tracking
   - Pass-by-pass status updates

6. **Completion & Certification**
   - Generate compliance certificates
   - Export audit documentation
   - Verify operation success

---

## API Reference

### Device Scanning
```javascript
const devices = await window.api.scanDevices();
// Returns: Array of device objects with metadata
```

### Wipe Operation
```javascript
await window.api.startWipe({
  device: '/dev/sdb',
  method: 'destroy',
  passes: 7
});
```

### Progress Monitoring
```javascript
window.api.onWipeProgress((progress) => {
  console.log(`Pass ${progress.pass}/${progress.totalPasses}`);
  console.log(`Progress: ${progress.percentage}%`);
});
```

---

## Development

### Building from Source
```bash
# Backend compilation
cd wipe-tool
make clean && make

# Frontend development
cd gui
npm install
npm run dev
```

### Testing
```bash
# Unit tests
npm test

# Security tests
npm run security-audit

# Cross-platform tests
npm run test:platforms
```

### Contributing Guidelines
1. Fork the repository
2. Create feature branch (`git checkout -b feature/enhancement`)
3. Test both C backend and Electron frontend
4. Run security audits
5. Submit pull request with detailed description

---

## Security Considerations

### ⚠️ Critical Warnings

- **Data Destruction**: This tool **permanently destroys** data. Recovery is impossible.
- **Device Verification**: Always verify target devices before execution.
- **Backup Requirements**: Ensure critical data is backed up elsewhere.
- **Elevated Privileges**: Root/Administrator access required for device-level operations.

### Best Practices
- Run in controlled, isolated environments
- Maintain audit trails for compliance
- Verify device identification before wiping
- Test operations on non-critical devices first
- Follow organizational data destruction policies

### Compliance Notes
- Designed for NIST 800-88 Rev. 1 standards
- Suitable for most regulatory requirements
- Verify local/industry-specific regulations
- Maintain certificates for audit purposes

---

## Troubleshooting

### Common Issues
| Issue | Solution |
|-------|----------|
| Device not detected | Run with elevated privileges |
| GUI won't start | Add `--no-sandbox` flag |
| Compilation errors | Install build-essential package |
| Permission denied | Execute as administrator/root |

### Support Resources
- Review component documentation in respective directories
- Test C tool independently: `./wipe-tool --help`
- Check system logs for detailed error messages
- Verify device permissions and accessibility

---

## License & Legal

**License**: MIT License - see [LICENSE](LICENSE) file for details.

**Compliance**: NIST 800-88 Rev. 1 compliant sanitization methods.

**Disclaimer**: Users are responsible for compliance with local regulations and proper data handling procedures. This tool is provided "as-is" without warranty.

*NullBytes Desktop App - Secure, Compliant, Professional Data Destruction*
