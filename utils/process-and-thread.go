package utils

import (
	file_and_directory "github.com/headend/share-module/file-and-directory"
	"log"
	"os"
	"strconv"
	"syscall"
)

func IsExistsProcess(pidStr string) (isExists bool, err error) {
	pid, err1 := strconv.ParseInt(pidStr, 10, 64)
	if err1 != nil {
		return false, err1
	}
	process, err2 := os.FindProcess(int(pid))
	if err != nil {
		log.Println(err)
		return false, err2
	}
	killErr := process.Signal(syscall.Signal(0))
	procExists := killErr == nil
	return procExists, nil
}


func IsWorkerRunning(pidPathFile string) bool {
	myFile := file_and_directory.MyFile{Path: pidPathFile}
	pidStr, err := myFile.Read()
	if err != nil {
		log.Println(err)
		return false
	} else {
		isExist, err3 := IsExistsProcess(pidStr)
		if err3 != nil {
			log.Println(err3)
			return false
		}
		if isExist {
			log.Println("Worker already running")
			return true
		}
	}
	return false
}