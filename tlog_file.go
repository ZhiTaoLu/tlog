package tlog

import (
	"fmt"
	"os"
	"time"
)

const logfilesize = 512 * 1024 * 1024

var logbasepath = "/home/dgd/kiseki_log"
var clearfile = false

type refreType int

const (
	date refreType = iota
	size
)

func SetClearFile(v bool) {
	clearfile = v
}

func SetLogsPath(path string) {
	logbasepath = path
}

func checkLogFile(typ logType, f *logFile) {
	if f == nil {
		refreshLogfile(typ)
		return
	}

	if !checkTime(f.creatTime) {
		refreshLogfile(typ)
		clearFile(f.creatTime, genPath(), pTaoLogSystem.logFileTyp[typ], 0)
		return
	}

	if checkSize(f.file) {
		refreshLogfile(typ)
	}
}

func checkSize(f *os.File) bool {
	info, _ := f.Stat()
	return info.Size() > logfilesize

}

func checkTime(cTime string) bool {
	return getCurrTime() == cTime
}

func refreshLogfile(typ logType) {
	oldFile := pTaoLogSystem.loggerFile[typ]
	if oldFile != nil {
		oldFile.file.Close()
	}

	path := genPath()
	fileName := logFileInit(path, pTaoLogSystem.logFileTyp[typ], 0)

	pLogFile, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return
	}

	pTaoLogSystem.loggerFile[typ] = &logFile{file: pLogFile, creatTime: getCurrTime()}
}

func getCurrTime() string {
	currTime := time.Now()
	y, m, d := currTime.Date()
	return fmt.Sprintf("%02d%02d%02d%02d", y, m, d, currTime.Hour())
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
	fileName := logDir + v + "-" + time.Now().Format("2006010215") + fmt.Sprintf("-%v", index) + ".log"
	b, _ := pathExists(fileName)
	if b {
		return logFileInit(logDir, v, index+1)
	} else {
		return fileName
	}

	return fileName
}

func genPath() string {
	logDir := fmt.Sprintf("%v/%v/", logbasepath, "inputs")
	if appName == "" {
		logDir = fmt.Sprintf("%v/%v/", logbasepath, "inputs")
	}

	if _, err := os.Stat(logDir); !(err == nil || os.IsExist(err)) {
		if merr := os.MkdirAll(logDir, os.ModeDir); merr != nil {
			fmt.Println(merr)
			return ""
		}
	}

	return logDir
}

func clearFile(creatTime, logDir, v string, index int) {
	handletime, _ := time.ParseInLocation("2006010215", creatTime, time.Local)
	fileName := logDir + v + "-" + handletime.Add(time.Hour*1*-1).Format("2006010215") + fmt.Sprintf("-%v", index) + ".log"
	b, _ := pathExists(fileName)
	if b {
		os.Remove(fileName)
		clearFile(creatTime, logDir, v, index+1)
	}
}
