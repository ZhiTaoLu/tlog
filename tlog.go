package tlog

import (
	"encoding/json"
	"fmt"
	"github.com/petermattis/goid"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type taoLogSystem struct {
	fLog       map[logType]*log.Logger
	loggerFile map[logType]*logFile
	logFileTyp map[logType]string
}

type logStru struct {
	//LoggerName string    `json:"logger_name"`
	Level   string `json:"level,omitempty"`
	AppName string `json:"app"`
	//Time    time.Time `json:"time"`
	TimeS   string                 `json:"time"`
	TimeU   float64                `json:"ts,omitempty"`
	GoID    string                 `json:"goid,omitempty"`
	Hint    string                 `json:"hint,omitempty"`
	Uid     string                 `json:"uid,omitempty"`
	Cid     string                 `json:"cid,omitempty"`
	Caller  string                 `json:"caller,omitempty"`
	Message string                 `json:"msg"`
	File    map[string]interface{} `json:"file,omitempty"`
}

type logFile struct {
	file      *os.File
	creatTime string
}

type outputLogstr struct {
	typ logType
	buf []byte
}

const (
	InfoLog  = infoLog
	ErrorLog = errorLog
)

type logType int8

const (
	infoLog logType = iota
	warningLog
	errorLog
)

const severityChar = "IWE"

var logLevel = map[logType]string{
	infoLog:  "info",
	errorLog: "error",
}

var pTaoLogSystem *taoLogSystem
var bMarkFile bool = false
var bWriteConsole bool = true
var once sync.Once
var outputChannel chan outputLogstr = make(chan outputLogstr, 10000)
var appName string

func Init(name string) {
	appName = name
	pTaoLogSystem = &taoLogSystem{
		//cLog:       make(map[logType]*log.Logger),
		fLog:       make(map[logType]*log.Logger),
		loggerFile: make(map[logType]*logFile),
		logFileTyp: make(map[logType]string),
	}

	pTaoLogSystem.logFileTyp[infoLog] = "Info"
	pTaoLogSystem.logFileTyp[errorLog] = "Err"

	if !pTaoLogSystem.genLogHandler() {
		fmt.Println("gen hander err")
		return
	}

	//outputChannel = make(chan outputLogstr, 1000)
	go outputChan()
}

func (this *taoLogSystem) genLogHandler() bool {
	logDir := ""

	if bMarkFile {
		logDir = genPath()
		if logDir == "" {
			fmt.Println("gen log path err")
			return false
		}
	}

	for k, v := range this.logFileTyp {
		if bMarkFile {
			fileName := logFileInit(logDir, v, 0)
			pLogFile, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				return false
			}

			if this.loggerFile[k] != nil {
				this.loggerFile[k].file.Close()
			}

			this.loggerFile[k] = &logFile{file: pLogFile, creatTime: getCurrTime()}
		}

		log2 := log.New(os.Stderr, "", 0)
		if log2 == nil {
			return false
		}
		this.fLog[k] = log2
	}

	return true
}

func (this *taoLogSystem) genGoId() string {
	return fmt.Sprintf("[%v,%v]", os.Getpid(), goid.Get())
}

func (this *taoLogSystem) genPrefix(typ logType) string {
	tTime := time.Now()
	strDate := fmt.Sprintf("%v%02d%02d", severityChar[typ:typ+1], tTime.Month(), tTime.Day())
	strTime := fmt.Sprintf("%02d:%02d:%02d.%v   ", tTime.Hour(), tTime.Minute(), tTime.Second(), tTime.Nanosecond()/1000)
	return fmt.Sprintf("%v %v [%v,%v] ", strDate, strTime, os.Getpid(), goid.Get())
}

func (this *taoLogSystem) genFlieLine() string {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		file = "???"
		line = 1
	} else {
		idx := strings.Index(file, "/src/")
		if idx >= 0 {
			file = file[idx+5:]
		}
	}

	tarLine := fmt.Sprintf("%v:%v", file, line)

	return tarLine
}

func SetWriteFile(bWrite bool) {
	bMarkFile = bWrite
}

func SetOutputTyp(bConsole, bWrite, bClear bool) {
	bWriteConsole = bConsole
	bMarkFile = bWrite
	clearfile = bClear
}

//func Info(args ...interface{}) {
//	//args = append(args, pMyLogSystem.genFlieLine())
//	outputPre(infoLog, fmt.Sprint(args...))
//}
//
//func Infoln(args ...interface{}) {
//	//args = append(args, pMyLogSystem.genFlieLine())
//	outputPre(infoLog, fmt.Sprintln(args...))
//}
//
//func Infof(format string, args ...interface{}) {
//	//args = append(args, pMyLogSystem.genFlieLine())
//	outputPre(infoLog, fmt.Sprintf(format, args...))ls
//}

func Infoj(msg string, args ...interface{}) {
	//args = append(args, pMyLogSystem.genFlieLine())
	outputPreJson(infoLog, msg, args...)
}

//func Error(args ...interface{}) {
//	//args = append(args, pMyLogSystem.genFlieLine())
//	outputPre(errorLog, fmt.Sprint(args...))
//}
//
//func Errorln(args ...interface{}) {
//	//args = append(args, pMyLogSystem.genFlieLine())
//	outputPre(errorLog, fmt.Sprintln(args...))
//}
//
//func Errorf(format string, args ...interface{}) {
//	//args = append(args, pMyLogSystem.genFlieLine())
//	outputPre(errorLog, fmt.Sprintf(format, args...))
//}

func Errorj(msg string, args ...interface{}) {
	//args = append(args, pMyLogSystem.genFlieLine())
	outputPreJson(errorLog, msg, args...)
}

func OutputPreHandle(typ logType, args ...interface{}) {
	files := make(map[string]interface{}, 0)
	if len(args) > 0 {
		for i := 0; i < len(args)-1; i += 2 {
			if i+1 > len(args) {
				files[fmt.Sprint(args[i])] = ""
				break
			}
			files[fmt.Sprint(args[i])] = args[i+1]
		}
	}

	bb, _ := json.Marshal(files)

	outputChannel <- outputLogstr{
		typ: typ,
		buf: bb,
	}
}

func OutputPre(typ logType, strlog string) {
	outputPre(typ, strlog)
}

func outputPre(typ logType, strlog string) {
	outputChannel <- outputLogstr{
		typ: typ,
		buf: []byte(strlog),
	}
}

func outputPreJson(typ logType, msg string, args ...interface{}) {
	files := make(map[string]interface{}, 0)
	if len(args) > 0 {
		for i := 0; i < len(args)-1; i += 2 {
			if i+1 > len(args) {
				files[fmt.Sprint(args[i])] = ""
				break
			}
			files[fmt.Sprint(args[i])] = args[i+1]
		}
	}

	currTime := time.Now()
	logF := logStru{
		//Level:   logLevel[typ],
		AppName: appName,
		TimeU:   float64(currTime.UnixNano()) / 1000000000,
		TimeS:   currTime.Format("2006-01-02T15:04:05.000+0800"),
		//GoID:    pTaoLogSystem.genGoId(),
		//Hint:    "",
		//Uid:     "",
		//Cid:     "",
		//Caller:  pTaoLogSystem.genFlieLine(),
		Message: msg,
		File:    files,
	}

	logFByt, _ := json.Marshal(logF)

	outputChannel <- outputLogstr{
		typ: typ,
		buf: logFByt,
	}
}

func outputChan() {
	for {
		select {
		case logstr := <-outputChannel:
			output(logstr.typ, string(logstr.buf))
		}
	}
}

func output(typ logType, logstr string) {
	pLog := pTaoLogSystem.fLog[typ]

	if bWriteConsole {
		pLog.SetOutput(os.Stderr)
		pLog.Output(2, logstr)
	}

	if bMarkFile {
		if typ != infoLog {
			checkLogFile(infoLog, pTaoLogSystem.loggerFile[infoLog])
			pLog.SetOutput(pTaoLogSystem.loggerFile[infoLog].file)
			pLog.Output(2, logstr)
		}

		checkLogFile(typ, pTaoLogSystem.loggerFile[typ])
		pLog.SetOutput(pTaoLogSystem.loggerFile[typ].file)
		pLog.Output(2, logstr)
	}
}
