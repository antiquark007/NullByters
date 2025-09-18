CrossPlatformDeskApp - NullBytes Go Desktop Application
This is the Go-based cross-platform desktop application for the NullBytes secure data wiping and certificate management suite. Built using Go and raylib, it provides a high-performance native GUI for NIST SP 800-88 compliant data sanitization with automated certificate generation and verification.
🎯 Overview
The CrossPlatformDeskApp is a native desktop application within the NullBytes suite, offering a user-friendly interface for secure data wiping and certificate management. It leverages raylib for cross-platform graphics and Go for efficient, reliable performance, supporting Windows, Linux, and macOS.
🚀 Features

Secure Data Wiping: Implements NIST SP 800-88 compliant Clear, Purge methods with zero fill, random fill, and cryptographic erase options.
Real-Time Device Detection: Automatically enumerates and identifies storage devices with system drive protection.
Certificate Management: Generates PDF/JSON certificates with digital signatures, QR code integration, and cloud upload capabilities.
Interactive GUI: Native raylib-based interface with real-time wiping progress visualization and certificate previews.
Cross-Platform Support: Runs natively on Windows, Linux, and macOS with consistent performance.
Compliance Logging: Maintains audit trails for regulatory compliance.

🛠️ Prerequisites
General Requirements

Go: Version 1.19 or later
Git: For cloning the repository

Platform-Specific Requirements
Linux

C compiler (e.g., gcc)

Raylib development libraries:
sudo apt update
sudo apt install -y build-essential git libgl1-mesa-dev libopenal-dev libx11-dev libxrandr-dev libxi-dev libxinerama-dev libxcursor-dev



Windows

MinGW-w64 compiler
Raylib Windows binaries

macOS

Xcode Command Line Tools
Raylib macOS framework

🚀 Installation and Setup
1. Clone the Repository
git clone https://github.com/antiquark007/NullBytes.git
cd NullBytes/CrossPlatformDeskApp

2. Install Go Dependencies
go mod tidy

3. Install Raylib Bindings
go get github.com/gen2brain/raylib-go/raylib

4. Platform-Specific Raylib Setup
Linux
git clone https://github.com/raysan5/raylib.git
cd raylib/src
make
sudo make install
cd ../..
sudo ldconfig

Windows

Download raylib-5.0_win64_mingw-w64.zip from raylib releases

Extract to C:\raylib

Set environment variables:
set CGO_CFLAGS=-IC:\raylib\include
set CGO_LDFLAGS=-LC:\raylib\lib -lraylib -lopengl32 -lgdi32 -lwinmm
set PATH=%PATH%;C:\raylib\lib


Copy raylib.dll to project directory or PATH


macOS
brew install raylib

Or build from source:
git clone https://github.com/raysan5/raylib.git
cd raylib/src
make PLATFORM=PLATFORM_DESKTOP
sudo make install
cd ../..

🏃‍♂️ Usage
Running the Application
Run in development mode:
go run ./cmd/app

Build and run for production:
Linux/macOS
go build -o nullbytes ./cmd/app
./nullbytes

Windows
go build -o nullbytes.exe ./cmd/app
nullbytes.exe

Main Interface

Dashboard: View detected devices and recent operations.
Wipe Operations: Select devices, choose NIST-compliant wiping methods, and monitor progress.
Certificate Management: Generate, sign, and upload compliance certificates with QR code support.
Settings: Customize preferences like window size and logging.

Example Operations
Device Wiping

Open Wipe tab
Select device from list
Choose method (Clear/Purge)
Configure passes and verification
Start wipe with confirmation
View progress and generate certificate

Certificate Generation

Go to Certificates tab
Select device and operation type
Choose PDF/JSON format
Enable QR code for mobile verification
Generate and optionally upload to cloud

📄 License
This project is licensed under the MIT License.