package prebuilt

import (
	_ "embed"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed sqlite3.exe
var sqlite3exe []byte

func ExtractSqlite3(projectDir string) (string, bool, error) {
	sqlite3, err := exec.LookPath("sqlite3")
	if err == nil {
		return sqlite3, false, nil
	}
	// TODO(nagi): support linux
	sqlite3Path := "." + string(os.PathSeparator) + filepath.Join(projectDir, "sqlite3.exe")
	sqlite3, err = exec.LookPath(sqlite3Path)
	if err == nil {
		return sqlite3, false, nil
	}
	err = os.WriteFile(sqlite3Path, sqlite3exe, 0755)
	if err != nil {
		log.Println("Failed to extract sqlite3.exe")
		return "", false, err
	}
	sqlite3, err = exec.LookPath(sqlite3Path)
	if err != nil {
		log.Println("Failed to lookup embed sqlite3")
		return "", true, err
	}
	return sqlite3, true, nil
}
