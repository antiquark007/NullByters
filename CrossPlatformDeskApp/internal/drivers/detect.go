package drivers

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)


type Drive struct {
	Name        string
	Path        string
	Type        string 
	FileSystem  string
	IsRemovable bool
	Device      string 
}


func GetDrives() ([]Drive, error) {
	switch runtime.GOOS {
	case "linux":
		return getLinuxDrives()
	case "windows":
		return getWindowsDrives()
	case "darwin":
		return getMacDrives()
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}


func getLinuxDrives() ([]Drive, error) {
	var drives []Drive
	

	mountedDrives := getMountedDrives()
	drives = append(drives, mountedDrives...)
	
	
	blockDevices := getBlockDevices()
	drives = append(drives, blockDevices...)
	
	
	commonMounts := getCommonMountPoints()
	drives = append(drives, commonMounts...)
	
	
	drives = removeDuplicateDrives(drives)
	
	return drives, nil
}


func getMountedDrives() []Drive {
	var drives []Drive
	
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return drives
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		
		device := fields[0]
		mountPoint := fields[1]
		fsType := fields[2]
		
		
		if strings.HasPrefix(mountPoint, "/sys") ||
			strings.HasPrefix(mountPoint, "/proc") ||
			strings.HasPrefix(mountPoint, "/dev") ||
			strings.HasPrefix(mountPoint, "/run") && !strings.Contains(mountPoint, "/media") ||
			mountPoint == "/" ||
			mountPoint == "/boot" ||
			strings.HasPrefix(device, "tmpfs") ||
			strings.HasPrefix(device, "devtmpfs") {
			continue
		}
		
		
		isUSB := false
		driveType := "internal"
		
		if strings.HasPrefix(device, "/dev/sd") || strings.HasPrefix(device, "/dev/nvme") {
			
			isUSB = isUSBDevice(device)
			if isUSB {
				driveType = "usb"
			}
		}
		
		
		name := filepath.Base(mountPoint)
		if name == "/" || name == "" {
			name = device
		}
		
	
		if strings.Contains(mountPoint, "/media/") || strings.Contains(mountPoint, "/mnt/") {
			parts := strings.Split(mountPoint, "/")
			if len(parts) > 0 {
				name = parts[len(parts)-1]
			}
		}
		
		drives = append(drives, Drive{
			Name:        name,
			Path:        mountPoint,
			Type:        driveType,
			FileSystem:  fsType,
			IsRemovable: isUSB,
			Device:      device,
		})
	}
	
	return drives
}


func getBlockDevices() []Drive {
	var drives []Drive
	
	
	cmd := exec.Command("lsblk", "-rno", "NAME,TYPE,SIZE,MOUNTPOINT,FSTYPE,MODEL,TRAN")
	output, err := cmd.Output()
	if err != nil {
		return drives
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		
		name := fields[0]
		devType := fields[1]
		mountPoint := fields[3]
		fsType := ""
		model := ""
		transport := ""
		
		if len(fields) > 4 {
			fsType = fields[4]
		}
		if len(fields) > 5 {
			model = fields[5]
		}
		if len(fields) > 6 {
			transport = fields[6]
		}
		
		
		if (devType != "part" && devType != "disk") || mountPoint == "" {
			continue
		}
		
		
		if mountPoint == "/" || mountPoint == "/boot" || strings.HasPrefix(mountPoint, "/boot/") {
			continue
		}
		
		
		isUSB := transport == "usb"
		driveType := "internal"
		if isUSB {
			driveType = "usb"
		}
		
		devicePath := "/dev/" + name
		
		
		driveName := name
		if model != "" && model != "N/A" {
			driveName = model
		} else if mountPoint != "" {
			driveName = filepath.Base(mountPoint)
		}
		
		drives = append(drives, Drive{
			Name:        driveName,
			Path:        mountPoint,
			Type:        driveType,
			FileSystem:  fsType,
			IsRemovable: isUSB,
			Device:      devicePath,
		})
	}
	
	return drives
}


func getCommonMountPoints() []Drive {
	var drives []Drive
	
	
	mountPoints := []string{
		"/media",
		"/mnt",
		fmt.Sprintf("/run/media/%s", os.Getenv("USER")),
		fmt.Sprintf("/media/%s", os.Getenv("USER")),
	}
	
	for _, base := range mountPoints {
		dirs, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		
		for _, d := range dirs {
			info, err := d.Info()
			if err != nil || !info.IsDir() {
				continue
			}
			
			fullPath := filepath.Join(base, d.Name())
			
		
			isUSB := false
			driveType := "external"
			
			
			if device := findDeviceForMountPoint(fullPath); device != "" {
				isUSB = isUSBDevice(device)
				if isUSB {
					driveType = "usb"
				}
			}
			
			drives = append(drives, Drive{
				Name:        d.Name(),
				Path:        fullPath,
				Type:        driveType,
				IsRemovable: isUSB,
			})
		}
	}
	
	return drives
}


func isUSBDevice(device string) bool {
	
	deviceName := strings.TrimPrefix(device, "/dev/")
	deviceName = strings.TrimRight(deviceName, "0123456789")
	
	
	sysPaths := []string{
		fmt.Sprintf("/sys/block/%s/removable", deviceName),
		fmt.Sprintf("/sys/class/block/%s/removable", deviceName),
	}
	
	for _, sysPath := range sysPaths {
		data, err := os.ReadFile(sysPath)
		if err == nil && strings.TrimSpace(string(data)) == "1" {
			return true
		}
	}
	
	
	usbPath := fmt.Sprintf("/sys/block/%s", deviceName)
	if linkDest, err := os.Readlink(usbPath); err == nil {
		if strings.Contains(linkDest, "/usb") {
			return true
		}
	}
	
	return false
}


func findDeviceForMountPoint(mountPoint string) string {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return ""
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && fields[1] == mountPoint {
			return fields[0]
		}
	}
	return ""
}


func getWindowsDrives() ([]Drive, error) {
	var drives []Drive
	
	
	cmd := exec.Command("wmic", "logicaldisk", "get", "size,freespace,name,volumename,drivetype,filesystem")
	output, err := cmd.Output()
	if err != nil {
		
		return getWindowsDrivesSimple()
	}
	
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return drives, nil
	}
	
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		
		
		driveType := "internal"
		isRemovable := false
		
		if len(fields) > 1 {
			switch fields[0] {
			case "2":
				driveType = "usb"
				isRemovable = true
			case "3":
				driveType = "internal"
			case "4":
				driveType = "network"
			case "5":
				driveType = "optical"
			}
		}
		
		
		driveLetter := ""
		volumeName := ""
		fileSystem := ""
		
		for _, field := range fields {
			if len(field) == 2 && field[1] == ':' {
				driveLetter = field
				break
			}
		}
		
		if driveLetter == "" {
			continue
		}
		
		
		if len(fields) > 3 {
			volumeName = fields[len(fields)-1]
		}
		
		name := driveLetter
		if volumeName != "" && volumeName != driveLetter {
			name = fmt.Sprintf("%s (%s)", volumeName, driveLetter)
		} else {
			name = fmt.Sprintf("Drive %s", driveLetter)
		}
		
		drives = append(drives, Drive{
			Name:        name,
			Path:        driveLetter + "\\",
			Type:        driveType,
			FileSystem:  fileSystem,
			IsRemovable: isRemovable,
			Device:      driveLetter,
		})
	}
	
	return drives, nil
}


func getWindowsDrivesSimple() ([]Drive, error) {
	var drives []Drive
	
	for letter := 'A'; letter <= 'Z'; letter++ {
		path := fmt.Sprintf("%c:\\", letter)
		if _, err := os.Stat(path); err == nil {
			drives = append(drives, Drive{
				Name:   fmt.Sprintf("Drive %c", letter),
				Path:   path,
				Type:   "unknown",
				Device: fmt.Sprintf("%c:", letter),
			})
		}
	}
	
	return drives, nil
}

func getMacDrives() ([]Drive, error) {
	var drives []Drive
	
	volumesPath := "/Volumes"
	dirs, err := os.ReadDir(volumesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read /Volumes: %w", err)
	}
	
	for _, d := range dirs {
		info, err := d.Info()
		if err != nil || !info.IsDir() {
			continue
		}
		
		fullPath := filepath.Join(volumesPath, d.Name())
		
		
		driveType := "external"
		isRemovable := false
		
		cmd := exec.Command("diskutil", "info", fullPath)
		if output, err := cmd.Output(); err == nil {
			outputStr := string(output)
			if strings.Contains(outputStr, "Protocol:") && strings.Contains(outputStr, "USB") {
				driveType = "usb"
				isRemovable = true
			}
		}
		
		drives = append(drives, Drive{
			Name:        d.Name(),
			Path:        fullPath,
			Type:        driveType,
			IsRemovable: isRemovable,
		})
	}
	
	return drives, nil
}


func removeDuplicateDrives(drives []Drive) []Drive {
	seen := make(map[string]bool)
	result := []Drive{}
	
	for _, drive := range drives {
		if drive.Path != "" && !seen[drive.Path] {
			seen[drive.Path] = true
			result = append(result, drive)
		}
	}
	
	return result
}