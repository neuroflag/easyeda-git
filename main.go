package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/neuroflag/easyeda-git/prebuilt"
)

var flagNoSplit = flag.Bool("no-split", false, "do not extract the sub documents into multiple sql files")
var flagNoStart = flag.Bool("no-start", false, "do not start the project after sync/open command")
var flagVerbose = flag.Bool("verbose", false, "print verbose messages")

func runSqlite3(sqlite3 string, dbPath string, command string) []byte {
	cmd := exec.Command(sqlite3, dbPath, command)
	if *flagVerbose {
		log.Println("Run sqlite3", cmd.Args)
	}
	cmd.Stderr = os.Stderr
	outputBytes, err := cmd.Output()
	if err != nil {
		log.Panic(err)
	}
	return outputBytes
}

func guessEprjPath(projectDir string, projectName string) (string, error) {
	for _, potentialFileName := range []string{
		projectName + ".eprj",
		"." + projectName + ".eprj",
	} {
		potentialFile := filepath.Join(projectDir, potentialFileName)
		_, err := os.Stat(potentialFile)
		if err == nil {
			return potentialFile, nil
		}
		if !os.IsNotExist(err) {
			log.Panic(err)
		}
	}
	return "", os.ErrNotExist
}

func commandSave(sqlite3 string, project string, projectDir string, projectName string) (bool, error) {
	eprjPath := ""
	if strings.HasSuffix(project, ".eprj") {
		eprjPath = project
		if _, err := os.Stat(eprjPath); err != nil && os.IsNotExist(err) {
			return false, err
		}
	} else {
		guessPath, err := guessEprjPath(projectDir, projectName)
		if err != nil {
			return false, err
		}
		eprjPath = guessPath
	}
	tmpSqlite3Path := filepath.Join(projectDir, projectName+".tmp.sqlite3")
	if *flagVerbose {
		log.Println("Clone", eprjPath, "to", tmpSqlite3Path)
	}
	projectSqliteBytes, err := os.ReadFile(eprjPath)
	if err != nil {
		log.Panic(err)
	}
	if err = os.WriteFile(tmpSqlite3Path, projectSqliteBytes, 0644); err != nil {
		log.Panic(err)
	}
	defer func() {
		os.Remove(tmpSqlite3Path)
	}()
	if !*flagNoSplit {
		documentsOutput := runSqlite3(sqlite3, tmpSqlite3Path,
			`SELECT "uuid", "display_title", "schematic_uuid", "project_uuid" FROM "documents"`)
		documentLines := strings.Split(strings.Trim(string(documentsOutput), " \r\n"), "\n")
		if *flagVerbose {
			log.Println("Found", len(documentLines), "documents")
		}
		processedTreeNames := []string{}
		for _, documentLine := range documentLines {
			documentMetadata := strings.Split(strings.Trim(documentLine, " \r\n"), "|")
			documentUuid := documentMetadata[0]
			displayTitle := documentMetadata[1]
			schematicUuid := documentMetadata[2]
			projectUuid := documentMetadata[3]
			if *flagVerbose {
				log.Println("Get tree name for", documentUuid, displayTitle, "(", projectUuid, schematicUuid, ")")
			}
			treeProjectNameOutput := runSqlite3(sqlite3, tmpSqlite3Path,
				fmt.Sprintf(`SELECT "name" FROM "projects" WHERE uuid = '%s'`, projectUuid))
			treeProjectName := strings.Trim(string(treeProjectNameOutput), " \r\n")
			treeSchematicName := ""
			if schematicUuid != "" {
				treeSchematicNameOutput := runSqlite3(sqlite3, tmpSqlite3Path,
					fmt.Sprintf(`SELECT "display_name" FROM "schematics" WHERE uuid = '%s'`, schematicUuid))
				treeSchematicName = strings.Trim(string(treeSchematicNameOutput), " \r\n")
			}
			treeName := fmt.Sprintf("%s_%s", treeProjectName, displayTitle)
			if schematicUuid != "" {
				treeName = fmt.Sprintf("%s_%s_%s", treeProjectName, treeSchematicName, displayTitle)
			}
			for _, processedTreeName := range processedTreeNames {
				if treeName == processedTreeName {
					log.Panic("Duplicate document name at ", treeName, " Please rename before continue")
				}
			}
			processedTreeNames = append(processedTreeNames, treeName)
			documentSqlPath := filepath.Join(projectDir, projectName+"_"+treeName+".eprj.sql")
			log.Println("Extract document", documentUuid, treeName, "into", documentSqlPath)
			dataStrOutput := runSqlite3(sqlite3, tmpSqlite3Path,
				fmt.Sprintf(`SELECT "dataStr" FROM "documents" WHERE uuid = '%s'`, documentUuid))
			dataStr := bytes.TrimRight(dataStrOutput, " \r\n")
			updateStatement := bytes.Join([][]byte{
				[]byte(`UPDATE "documents" SET "dataStr" = '`),
				dataStr,
				[]byte(fmt.Sprintf(`' WHERE "uuid" = '%s';`, documentUuid)),
			}, []byte{})
			if err := os.WriteFile(documentSqlPath, updateStatement, 0644); err != nil {
				log.Panic(err)
			}
			runSqlite3(sqlite3, tmpSqlite3Path,
				fmt.Sprintf(`UPDATE "documents" SET "dataStr" = '' WHERE uuid = '%s'`, documentUuid))
		}
	}
	dump := runSqlite3(sqlite3, tmpSqlite3Path, ".dump")
	eprjSqlPath := filepath.Join(projectDir, projectName+".eprj.sql")
	isInitEprjSql := false
	if _, err = os.Stat(eprjSqlPath); err != nil && os.IsNotExist(err) {
		isInitEprjSql = true
	}
	if err := os.WriteFile(eprjSqlPath, dump, 0644); err != nil {
		log.Panic(err)
	}
	return isInitEprjSql, nil
}

func commandOpen(sqlite3 string, project string, projectDir string, projectName string) string {
	eprjSqlPath := ""
	if strings.HasSuffix(project, ".eprj.sql") {
		eprjSqlPath = project
		_, err := os.Stat(eprjSqlPath)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Panic(err)
			}
		}
	} else {
		eprjSqlPath = filepath.Join(projectDir, projectName+".eprj.sql")
	}
	eprjPath := filepath.Join(projectDir, projectName+".eprj")
	_, err := os.Stat(eprjPath)
	if err == nil || !os.IsNotExist(err) {
		backupDir := filepath.Join(projectDir, projectName+"_backup")
		if *flagVerbose {
			log.Println("Creating backup directory", backupDir)
		}
		err := os.MkdirAll(backupDir, 0755)
		if err != nil {
			log.Panic(err)
		}
		eprjBackupPath := filepath.Join(projectDir, projectName+"_backup",
			projectName+time.Now().Format("20060102150405")+".eprj")
		log.Println("Moving", eprjPath, "to", backupDir)
		err = os.Rename(eprjPath, eprjBackupPath)
		if err != nil {
			log.Panic(err)
		}
	}
	log.Println("Load", eprjSqlPath)
	runSqlite3(sqlite3, eprjPath, fmt.Sprintf(`.read %q`, eprjSqlPath))
	projectDirFiles, err := os.ReadDir(projectDir)
	if err != nil {
		log.Panic(err)
	}
	for _, projectDirFile := range projectDirFiles {
		if strings.HasPrefix(projectDirFile.Name(), projectName+"_") &&
			strings.HasSuffix(projectDirFile.Name(), ".eprj.sql") {
			documentSqlPath := filepath.Join(projectDir, projectDirFile.Name())
			log.Println("Load", documentSqlPath)
			runSqlite3(sqlite3, eprjPath, fmt.Sprintf(`.read %q`, documentSqlPath))
		}
	}
	return eprjPath
}

func startEasyEda(eprjPath string) {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", fmt.Sprintf("/C start %v", eprjPath))
		log.Println("Start", eprjPath)
		if *flagVerbose {
			log.Println("Run", cmd.Args)
		}
		if err := cmd.Start(); err != nil {
			log.Panic(err)
		}
	} else if runtime.GOOS == "linux" {
		if _, err := os.Stat("/opt/easyeda-pro/easyeda-pro"); err == nil {
			cmd := exec.Command("/opt/easyeda-pro/easyeda-pro", eprjPath)
			log.Println("Start", eprjPath)
			if *flagVerbose {
				log.Println("Run", cmd.Args)
			}
			if err := cmd.Start(); err != nil {
				log.Panic(err)
			}
		} else if _, err := os.Stat("/opt/lceda-pro/lceda-pro"); err == nil {
			cmd := exec.Command("/opt/lceda-pro/lceda-pro", eprjPath)
			log.Println("Start", eprjPath)
			if *flagVerbose {
				log.Println("Run", cmd.Args)
			}
			if err := cmd.Start(); err != nil {
				log.Panic(err)
			}
		} else {
			log.Panic("EasyEDA / LCEDA is not found in /opt")
		}
	} else {
		log.Panic("Unsupported OS", runtime.GOOS)
	}
}

func main() {
	flag.BoolVar(flagVerbose, "v", false, "same as -verbose")
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintln(w, "Usage: easyeda-git [sync]    <eprj-file>")
		fmt.Fprintln(w, "       easyeda-git save|open <eprj-file>")
		fmt.Fprintln(w, "       lceda-git   [sync]    <eprj-file>")
		fmt.Fprintln(w, "       lceda-git   save|open <eprj-file>")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "Convert easyeda pro sqlite project file into sql code for eaiser diff")
		fmt.Fprintln(w, "and version control")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "Example:")
		fmt.Fprintln(w, "       $ easyeda-git MyProject.eprj")
		fmt.Fprintln(w, "               This will run the sync command, which does two things together:")
		fmt.Fprintln(w, "                   1. `save` - convert eprj file into sql files")
		fmt.Fprintln(w, "                   2. `open` - restore eprj file from sql files")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "       If you prefer to run separately, two subcommands are also provided")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "       $ easyeda-git save MyProject.eprj")
		fmt.Fprintln(w, "               Convert eprj file into sql files")
		fmt.Fprintln(w, "               This is usually done before committing the changes")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "       $ easyeda-git open MyProject.eprj.sql")
		fmt.Fprintln(w, "               Restore eprj file from sql files, start EasyEDA GUI")
		fmt.Fprintln(w, "               This is usually done after checking out the source code")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "Options:")
		flag.PrintDefaults()
	}
	flag.Parse()
	command := flag.Arg(0)
	project := flag.Arg(1)
	if command != "save" && command != "open" && command != "sync" {
		project = command
		command = "sync"
	}
	if *flagVerbose {
		log.Println("Command:", command)
		log.Println("Project:", project)
	}
	if project == "" {
		log.Panic("usage: easyeda-git [sync|save|open] <eprj-file>")
	}
	projectDir := filepath.Dir(project)
	projectName := filepath.Base(project)
	projectName = strings.TrimSuffix(projectName, ".sql")
	projectName = strings.TrimSuffix(projectName, ".eprj")
	projectName = strings.TrimPrefix(projectName, ".")
	sqlite3, created, err := prebuilt.ExtractSqlite3(projectDir)
	if created {
		if *flagVerbose {
			log.Println("Extract sqlite3 at", sqlite3)
		}
		defer func() {
			if *flagVerbose {
				log.Println("Removing", sqlite3)
			}
			os.Remove(sqlite3)
		}()
	}
	if err != nil {
		log.Panic(err)
	}
	if *flagVerbose && !created {
		log.Println("Found sqlite3 at", sqlite3)
	}
	if command == "save" {
		isInitEprjSql, err := commandSave(sqlite3, project, projectDir, projectName)
		if err != nil {
			log.Panic(err)
		}
		if isInitEprjSql {
			eprjPath := commandOpen(sqlite3, project, projectDir, projectName)
			if _, err := commandSave(sqlite3, eprjPath, projectDir, projectName); err != nil {
				log.Panic(err)
			}
		}
	} else if command == "open" {
		eprjPath := commandOpen(sqlite3, project, projectDir, projectName)
		if !*flagNoStart {
			startEasyEda(eprjPath)
		}
	} else if command == "sync" {
		isInitEprjSql, err := commandSave(sqlite3, project, projectDir, projectName)
		if err != nil && !os.IsNotExist(err) {
			log.Panic(err)
		}
		if isInitEprjSql {
			eprjPath := commandOpen(sqlite3, project, projectDir, projectName)
			if _, err := commandSave(sqlite3, eprjPath, projectDir, projectName); err != nil {
				log.Panic(err)
			}
		}
		eprjPath := commandOpen(sqlite3, project, projectDir, projectName)
		if !*flagNoStart {
			startEasyEda(eprjPath)
		}
	}
}
