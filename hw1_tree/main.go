package main

import (
	"os"
	"io/ioutil"
	"fmt"
	"io"
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

func dirTree(file io.Writer, path string, printFiles bool) error {
	return dirTreeWithPrefix(file, path, printFiles, "")
}

func dirTreeWithPrefix(file io.Writer, path string, printFiles bool, prefix string) error {
	fileInfos, _ := ioutil.ReadDir(path)
	if !printFiles {
		fileInfosFilterded := fileInfos[:0]
		for _, fileInfo := range fileInfos {
			if fileInfo.IsDir() {
				fileInfosFilterded = append(fileInfosFilterded, fileInfo)
			}
		}
		fileInfos = fileInfosFilterded
	}
	start := prefix + "├───"
	cont := prefix + "│	"
	for i, fileInfo := range fileInfos {
		if i == len(fileInfos)-1 {
			start = prefix + "└───"
			cont = prefix + "	"
		}
		size := ""
		if !fileInfo.IsDir() {
			if fileInfo.Size() == 0 {
				size = " (empty)"
			} else {
				size = fmt.Sprintf(" (%db)", fileInfo.Size())
			}
		}
		fmt.Fprintln(file, start+fileInfo.Name()+size)
		if fileInfo.IsDir() {
			dirTreeWithPrefix(file, path + string(os.PathSeparator) + fileInfo.Name(), printFiles, cont)
		}
	}
	return nil
}
