package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	dirPath    = flag.String("d", "", "Directory to publish")
	branchName = flag.String("b", "", "Branch to publish to")
)

const dTmpDirName = "go-gh-page-tool-temp"

func copy(src, dst string) error {
	if strings.Contains(src, dTmpDirName) {
		return nil
	}
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst, srcInfo)
}

func copyFile(src, dst string, srcInfo os.FileInfo) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	if _, err = io.Copy(destination, source); err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}

func copyDir(srcDir, dstDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			return err
		}
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if err := copy(srcPath, dstPath); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	flag.Parse()

	if len(flag.Args()) == 0 || flag.Arg(0) != "publish" {
		fmt.Println("Usage: ./go_gh_page_tool publish -d <directory> -b <branch>")
		os.Exit(1)
	}

	tmpDirPath := filepath.Join(*dirPath, dTmpDirName)
	if _, err := os.Stat(tmpDirPath); os.IsNotExist(err) {
		err := os.Mkdir(tmpDirPath, 0755)
		if err != nil {
			fmt.Printf("Error creating directory: %s\n", err)
			return
		}
	}

	err := copy(*dirPath, tmpDirPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		err := os.RemoveAll(tmpDirPath)
		if err != nil {
			fmt.Printf("Error removing directory: %s\n", err)
		}
	}()

	cmd := exec.Command("git", "checkout", "-b", *branchName)
	cmd.Dir = tmpDirPath
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error checkout:", err)
		return
	}

	cmd = exec.Command("git", "push", "-f", "origin", *branchName)
	cmd.Dir = tmpDirPath
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error push:", err)
		return
	}
}
