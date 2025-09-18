package drivers

import "os"

func GetDriveContents(path string) ([]string, error) {
	entries := []string{}
	files, err := os.ReadDir(path)
	if err != nil {
		return entries, err
	}
	for _, f := range files {
		entries = append(entries, f.Name())
	}
	return entries, nil
}
