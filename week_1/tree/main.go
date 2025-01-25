package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, path string, printFiles bool) error {

	err := walk(out, path, printFiles, "", 0)

	if err != nil {
		return fmt.Errorf("error walk: %v", err)
	}

	return nil
}

func walk(out io.Writer, path string, printFiles bool, indentation string, level int) error {
	vertical := "│"
	tab := "\t"
	lastElIndentation := "└───"
	elIndentation := "├───"
	files, err := os.ReadDir(path)

	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	if !printFiles {
		files = filterDirr(files)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	lenFiles := len(files)

	for i, file := range files {
		el := ""
		size := ""
		if i == lenFiles-1 {
			el = lastElIndentation
		} else {
			el = elIndentation
		}

		size, err = getFormatedFileSize(file)

		if err != nil {
			return fmt.Errorf("get file size error %v", err)
		}
		fmt.Fprintln(out, indentation+el+file.Name()+size)

		if file.IsDir() {
			newLevel := level + 1
			if i == lenFiles-1 {
				vertical = ""
			}
			newIndentation := indentation + vertical + tab
			walk(out, path+"/"+file.Name(), printFiles, newIndentation, newLevel)
		}
	}
	return nil
}

func filterDirr(files []os.DirEntry) []os.DirEntry {
	acc := []os.DirEntry{}
	for _, file := range files {
		if file.IsDir() {
			acc = append(acc, file)
		}
	}
	return acc
}

func getFormatedFileSize(file os.DirEntry) (string, error) {
	if file.IsDir() {
		return "", nil
	}
	info, err := file.Info()

	if err != nil {
		return "", err
	}
	size := ""

	if info.Size() == 0 {
		size = "empty"
	} else {
		size = strconv.FormatInt(info.Size(), 10) + "b"
	}

	return " " + "(" + size + ")", nil
}
