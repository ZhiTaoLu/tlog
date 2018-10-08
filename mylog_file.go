package mlog

import (
	"os"
	"fmt"
	"time"
)

const logfilesize = 512 * 1024 * 1024

type refreType int

const (
	date refreType = iota
	size
)

func checkLogFile(typ logType, f *logFile) {
	if !checkTime(f.creatTime) {
		refreshLogfile(date, typ)
		return
	}

	if checkSize(f.file) {
		refreshLogfile(size, typ)
	}
}

func checkSize(f *os.File) bool {
	info, _ := f.Stat()
	return info.Size() > logfilesize

}

func checkTime(cTime string) bool {
	return getCurrTime() == cTime
}

func refreshLogfile(rTyp refreType, typ logType) {
	oldFile := pMyLogSystem.loggerFile[typ]
	if oldFile != nil {
		oldFile.file.Close()
	}

	path := genPath()
	fileName := logFileInit(path, pMyLogSystem.logFileTyp[typ], 0)

	pLogFile, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return
	}

	pMyLogSystem.loggerFile[typ] = &logFile{file: pLogFile, creatTime: getCurrTime()}
}

func getCurrTime() string {
	currTime := time.Now()
	y, m, d := currTime.Date()
	return fmt.Sprintf("%02d%02d%02d", y, m, d)
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func logFileInit(logDir, v string, index int) string {
	fileName := logDir + v + "-" + time.Now().Format("20060102") + fmt.Sprintf("-%v", index) + ".log"
	b, _ := pathExists(fileName)
	if b {
		return logFileInit(logDir, v, index+1)
	} else {
		return fileName
	}

	return fileName
}

func genPath() string {
	t := time.Now()
	logDir := "syslog/"
	if _, err := os.Stat(logDir); !(err == nil || os.IsExist(err)) {
		if merr := os.Mkdir(logDir, os.ModeDir); merr != nil {
			return ""
		}
	}

	logDir += t.Format("2006-01/")
	if _, err := os.Stat(logDir); !(err == nil || os.IsExist(err)) {
		if merr := os.Mkdir(logDir, os.ModeDir); merr != nil {
			return ""
		}
	}

	return logDir
}
