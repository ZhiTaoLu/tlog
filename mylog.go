package mlog

import (
	"log"
	"os"
	"time"
	"fmt"
	"runtime"
	"strings"
)

var bMarkFile bool = true

type logType int8

const (
	infoLog    logType = iota
	warningLog
	errorLog
)

const severityChar = "IWE"

type myLogSystem struct {
	fLog       map[logType]*log.Logger
	loggerFile map[logType]*logFile
	logFileTyp map[logType]string
}

type logFile struct {
	file      *os.File
	creatTime string
}

var pMyLogSystem *myLogSystem

func init() {
	pMyLogSystem = &myLogSystem{
		//cLog:       make(map[logType]*log.Logger),
		fLog:       make(map[logType]*log.Logger),
		loggerFile: make(map[logType]*logFile),
		logFileTyp: make(map[logType]string),
	}

	pMyLogSystem.logFileTyp[infoLog] = "Info"
	pMyLogSystem.logFileTyp[errorLog] = "Err"

	if !pMyLogSystem.genLogHandler() {
		return
	}
}

func (this *myLogSystem) genLogHandler() bool {
	logDir := genPath()
	if logDir == "" {
		return false
	}

	for k, v := range this.logFileTyp {
		fileName := logFileInit(logDir, v, 0)

		//if bMarkFile {
		pLogFile, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return false
		}

		if this.loggerFile[k] != nil {
			this.loggerFile[k].file.Close()
		}

		this.loggerFile[k] = &logFile{file: pLogFile, creatTime: getCurrTime()}
		//}

		log2 := log.New(os.Stderr, "", 0)
		if log2 == nil {
			return false
		}
		this.fLog[k] = log2
	}

	return true
}

func (this *myLogSystem) genPrefix(typ logType) string {
	tTime := time.Now()
	strDate := fmt.Sprintf("%v%02d%02d", severityChar[typ:typ+1], tTime.Month(), tTime.Day())
	strTime := fmt.Sprintf("%02d:%02d:%02d.%v   ", tTime.Hour(), tTime.Minute(), tTime.Second(), tTime.Nanosecond()/1000)
	return fmt.Sprintf("%v %v [%v] ", strDate, strTime, os.Getpid())
}

func (this *myLogSystem) genFlieLine() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 1
	} else {
		idx := strings.Index(file, "/src/")
		if idx >= 0 {
			file = file[idx+5:]
		}
	}

	tarLine := fmt.Sprintf("   - %v:%v", file, line)

	return tarLine
}

func SetWriteFile(bWrite bool) {
	bMarkFile = bWrite
}

func Info(args ...interface{}) {
	args = append(args, pMyLogSystem.genFlieLine())
	output(infoLog, fmt.Sprint(args...))
}

func Infoln(args ...interface{}) {
	args = append(args, pMyLogSystem.genFlieLine())
	output(infoLog, fmt.Sprintln(args...))
}

func Infof(format string, args ...interface{}) {
	args = append(args, pMyLogSystem.genFlieLine())
	output(infoLog, fmt.Sprintf(format+"%v", args...))
}

func Error(args ...interface{}) {
	args = append(args, pMyLogSystem.genFlieLine())
	output(errorLog, fmt.Sprint(args...))
}

func Errorln(args ...interface{}) {
	args = append(args, pMyLogSystem.genFlieLine())
	output(errorLog, fmt.Sprintln(args...))
}

func Errorf(format string, args ...interface{}) {
	args = append(args, pMyLogSystem.genFlieLine())
	output(errorLog, fmt.Sprintf(format+"%v", args...))
}

func output(typ logType, strlog string) {
	pLog := pMyLogSystem.fLog[typ]

	pPrefix := pMyLogSystem.genPrefix(typ)
	pLog.SetPrefix(pPrefix)
	pLog.SetOutput(os.Stderr)
	pLog.Output(2, strlog)

	if bMarkFile {
		if typ != infoLog {
			checkLogFile(infoLog, pMyLogSystem.loggerFile[infoLog])
			pLog.SetOutput(pMyLogSystem.loggerFile[infoLog].file)
			pLog.Output(2, strlog)
		}

		checkLogFile(typ, pMyLogSystem.loggerFile[typ])
		pLog.SetOutput(pMyLogSystem.loggerFile[typ].file)
		pLog.Output(2, strlog)
	}
}
