package rclone

import "os"

func ListFolders(root string) ([]string, error) {

	var folders []string

	dirs, err := os.ReadDir(root)

	if err != nil {
		return nil, err
	}

	for _, dir := range dirs {
		if dir.IsDir() {
			folders = append(folders, dir.Name())
		}
	}

	return folders, nil
}
