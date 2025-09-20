package drivers

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ClearItem performs basic file/directory deletion
// This is a standard delete operation that removes files/directories from the filesystem
func ClearItem(path string) error {
	if path == "" {
		return errors.New("path cannot be empty")
	}

	// Check if path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}
	if err != nil {
		return fmt.Errorf("failed to stat path %s: %v", path, err)
	}

	// Safety check - prevent deletion of critical system paths
	if isCriticalPath(path) {
		return fmt.Errorf("cannot delete critical system path: %s", path)
	}

	if info.IsDir() {
		// Remove directory and all its contents
		err = os.RemoveAll(path)
		if err != nil {
			return fmt.Errorf("failed to remove directory %s: %v", path, err)
		}
		fmt.Printf("Directory cleared: %s\n", path)
	} else {
		// Remove single file
		err = os.Remove(path)
		if err != nil {
			return fmt.Errorf("failed to remove file %s: %v", path, err)
		}
		fmt.Printf("File cleared: %s\n", path)
	}

	return nil
}

// PurgeItem performs secure deletion with multiple overwrite passes
// This function attempts to securely overwrite data before deletion
func PurgeItem(path string) error {
	if path == "" {
		return errors.New("path cannot be empty")
	}

	// Check if path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}
	if err != nil {
		return fmt.Errorf("failed to stat path %s: %v", path, err)
	}

	// Safety check - prevent purging of critical system paths
	if isCriticalPath(path) {
		return fmt.Errorf("cannot purge critical system path: %s", path)
	}

	if info.IsDir() {
		// Recursively purge directory contents
		err = purgeDirectory(path)
		if err != nil {
			return fmt.Errorf("failed to purge directory %s: %v", path, err)
		}
		fmt.Printf("Directory purged: %s\n", path)
	} else {
		// Purge single file
		err = purgeFile(path)
		if err != nil {
			return fmt.Errorf("failed to purge file %s: %v", path, err)
		}
		fmt.Printf("File purged: %s\n", path)
	}

	return nil
}

// purgeFile securely overwrites and deletes a single file
func purgeFile(filePath string) error {
	// Try platform-specific secure deletion tools first
	if err := trySecureDeleteTool(filePath); err == nil {
		return nil
	}

	// Fallback to manual overwrite if tools aren't available
	return manualSecureDelete(filePath)
}

// trySecureDeleteTool attempts to use OS-specific secure deletion tools
func trySecureDeleteTool(filePath string) error {
	switch runtime.GOOS {
	case "linux":
		// Try shred command (most common on Linux)
		if _, err := exec.LookPath("shred"); err == nil {
			cmd := exec.Command("shred", "-vfz", "-n", "3", "-u", filePath)
			return cmd.Run()
		}

		// Try wipe command as alternative
		if _, err := exec.LookPath("wipe"); err == nil {
			cmd := exec.Command("wipe", "-rf", filePath)
			return cmd.Run()
		}

	case "darwin":
		// Try rm with secure deletion on macOS
		cmd := exec.Command("rm", "-P", filePath)
		if err := cmd.Run(); err == nil {
			return nil
		}

	case "windows":
		// Try sdelete if available (Sysinternals tool)
		if _, err := exec.LookPath("sdelete"); err == nil {
			cmd := exec.Command("sdelete", "-p", "3", "-s", "-z", filePath)
			return cmd.Run()
		}

		// Try cipher command (built into Windows)
		if _, err := exec.LookPath("cipher"); err == nil {
			// First delete the file normally
			os.Remove(filePath)
			// Then overwrite free space in the directory
			dir := filepath.Dir(filePath)
			cmd := exec.Command("cipher", "/w:"+dir)
			return cmd.Run()
		}
	}

	return errors.New("no secure deletion tool available")
}

// manualSecureDelete performs manual secure deletion with multiple overwrite passes
func manualSecureDelete(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open file for overwriting: %v", err)
	}
	defer file.Close()

	// Get file size
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}
	fileSize := info.Size()

	// Perform multiple overwrite passes
	passes := [][]byte{
		// Pass 1: All zeros
		make([]byte, fileSize),
		// Pass 2: All ones (0xFF)
		func() []byte {
			data := make([]byte, fileSize)
			for i := range data {
				data[i] = 0xFF
			}
			return data
		}(),
		// Pass 3: Random data
		func() []byte {
			data := make([]byte, fileSize)
			rand.Read(data)
			return data
		}(),
	}

	for passNum, passData := range passes {
		// Seek to beginning of file
		_, err := file.Seek(0, 0)
		if err != nil {
			return fmt.Errorf("failed to seek to beginning on pass %d: %v", passNum+1, err)
		}

		// Write the pattern
		_, err = file.Write(passData)
		if err != nil {
			return fmt.Errorf("failed to write overwrite data on pass %d: %v", passNum+1, err)
		}

		// Force write to disk
		err = file.Sync()
		if err != nil {
			return fmt.Errorf("failed to sync on pass %d: %v", passNum+1, err)
		}

		fmt.Printf("Completed overwrite pass %d/%d for %s\n", passNum+1, len(passes), filePath)
	}

	file.Close()

	// Finally delete the file
	err = os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("failed to delete file after overwriting: %v", err)
	}

	return nil
}

// purgeDirectory recursively purges all files in a directory
func purgeDirectory(dirPath string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself on first pass
		if path == dirPath && info.IsDir() {
			return nil
		}

		if info.IsDir() {
			// For subdirectories, we'll handle them after their contents are processed
			return nil
		} else {
			// Purge individual files
			return purgeFile(path)
		}
	})
}

// isCriticalPath checks if a path is critical and should not be deleted
func isCriticalPath(path string) bool {
	// Normalize path
	path = filepath.Clean(path)

	// Define critical paths for different operating systems
	criticalPaths := map[string][]string{
		"linux": {
			"/", "/bin", "/sbin", "/usr", "/etc", "/lib", "/lib64",
			"/boot", "/sys", "/proc", "/dev", "/run", "/var/log",
		},
		"windows": {
			"C:\\Windows", "C:\\Program Files", "C:\\Program Files (x86)",
			"C:\\System Volume Information", "C:\\PerfLogs",
		},
		"darwin": {
			"/System", "/Library", "/usr", "/bin", "/sbin", "/etc",
			"/Applications", "/var", "/tmp",
		},
	}

	// Check against OS-specific critical paths
	if paths, exists := criticalPaths[runtime.GOOS]; exists {
		for _, criticalPath := range paths {
			// Check if the path is exactly a critical path or starts with it
			if strings.EqualFold(path, criticalPath) ||
				strings.HasPrefix(strings.ToLower(path), strings.ToLower(criticalPath)+string(os.PathSeparator)) {
				return true
			}
		}
	}

	// Additional safety checks
	if path == "/" || path == "C:\\" || path == "" {
		return true
	}

	return false
}

// GetSecureDeleteCapabilities returns information about available secure deletion methods
func GetSecureDeleteCapabilities() map[string]bool {
	capabilities := make(map[string]bool)

	switch runtime.GOOS {
	case "linux":
		_, err := exec.LookPath("shred")
		capabilities["shred"] = (err == nil)

		_, err = exec.LookPath("wipe")
		capabilities["wipe"] = (err == nil)

	case "darwin":
		_, err := exec.LookPath("rm")
		capabilities["rm_secure"] = (err == nil)

	case "windows":
		_, err := exec.LookPath("sdelete")
		capabilities["sdelete"] = (err == nil)

		_, err = exec.LookPath("cipher")
		capabilities["cipher"] = (err == nil)
	}

	capabilities["manual_overwrite"] = true // Always available
	return capabilities
}

// Helper function to estimate purge time based on file size
func EstimatePurgeTime(filePath string) (int, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}

	// Rough estimate: 3 passes * file size / average write speed (50 MB/s)
	fileSizeMB := info.Size() / (1024 * 1024)
	estimatedSeconds := int((fileSizeMB * 3) / 50)

	if estimatedSeconds < 1 {
		estimatedSeconds = 1
	}

	return estimatedSeconds, nil
}
