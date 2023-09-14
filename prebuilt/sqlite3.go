package prebuilt

import (
	_ "embed"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

//go:embed sqlite3-win32-x86.exe
var sqlite3_win32_x86 []byte

//go:embed sqlite3-osx-x86
var sqlite3_osx_x86 []byte

//go:embed sqlite3-linux-x86
var sqlite3_linux_x86 []byte

func ExtractSqlite3(projectDir string) (string, bool, error) {
	sqlite3, err := exec.LookPath("sqlite3")
	if err == nil {
		return sqlite3, false, nil
	}
	sqlite3Path := ""
	if runtime.GOOS == "windows" && (runtime.GOARCH == "386" || runtime.GOARCH == "amd64") {
		sqlite3Path = "." + string(os.PathSeparator) + filepath.Join(projectDir, "sqlite3.exe")
	} else if runtime.GOOS == "darwin" {
		sqlite3Path = "." + string(os.PathSeparator) + filepath.Join(projectDir, "sqlite3")
	} else if runtime.GOOS == "linux" && (runtime.GOARCH == "386" || runtime.GOARCH == "amd64") {
		sqlite3Path = "." + string(os.PathSeparator) + filepath.Join(projectDir, "sqlite3")
	} else {
		log.Panic("Unsupported OS/ARCH", runtime.GOOS, runtime.GOARCH)
	}
	sqlite3, err = exec.LookPath(sqlite3Path)
	if err == nil {
		return sqlite3, false, nil
	}
	if runtime.GOOS == "windows" && (runtime.GOARCH == "386" || runtime.GOARCH == "amd64") {
		err = os.WriteFile(sqlite3Path, sqlite3_win32_x86, 0755)
	} else if runtime.GOOS == "darwin" {
		err = os.WriteFile(sqlite3Path, sqlite3_osx_x86, 0755)
	} else if runtime.GOOS == "linux" && (runtime.GOARCH == "386" || runtime.GOARCH == "amd64") {
		err = os.WriteFile(sqlite3Path, sqlite3_linux_x86, 0755)
	} else {
		log.Panic("Unsupported OS/ARCH", runtime.GOOS, runtime.GOARCH)
	}
	if err != nil {
		log.Println("Failed to extract sqlite3 excutable")
		return "", false, err
	}
	sqlite3, err = exec.LookPath(sqlite3Path)
	if err != nil {
		log.Println("Failed to lookup embed sqlite3")
		return "", true, err
	}
	return sqlite3, true, nil
}
