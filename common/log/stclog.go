package log

import (
	"os"
	slog"log"
	"strings"
	"fmt"
	"runtime"
	"io"
	"time"
	"starchain/common/config"
)
const(
	MAXSIZE = 10240
)
const(
	TRACE = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
)
var l *stclog
var logger *slog.Logger

func init(){
	//logPath := "./Logs/"
	logName := "stc.log"
	logLevel := config.Parameters.PrintLevel
	logPath := config.Parameters.LogPath
	if logPath == ""{
		logPath = "./Logs/"
	}
	if !strings.HasSuffix(logPath,"/"){
		logPath = logPath+"/"
	}
	//var writes []io.Writer
	_,err := os.Stat(logPath)
	if err != nil && os.IsNotExist(err){
		err := os.Mkdir(logPath,os.ModePerm)
		if err != nil{
			slog.Fatal("mkeir Logs folder failed")
		}
	}
	logfile,err := os.OpenFile(logPath+logName,os.O_APPEND|os.O_CREATE,os.ModePerm)
	if err != nil{
		//slog.Println(err)
		//return
		slog.Fatal(err)
	}
	fmt.Println(logPath+logName)
	//writes = append(writes,logfile)
	//writes = append(writes,os.Stdout)
	//writes := []io.Writer{os.Stdout,logfile}
	logio := io.MultiWriter(os.Stdout,logfile)
	logger = slog.New(logio,"",slog.Ldate|slog.Ltime)
	l = &stclog{c:make(chan string,MAXSIZE),level:getLevel(logLevel)}
	rotateLog(func() {
		logfile.Close()
		d, _ := time.ParseDuration("-24h")
		err = os.Rename(logPath+logName,logPath+"stc-"+time.Now().Add(d).Format("2006-01-02")+".log")
		if err != nil {
			slog.Println(err)
		}
		logfile,err = os.OpenFile(logPath+logName,os.O_APPEND|os.O_CREATE,os.ModePerm)
		if err != nil {
			slog.Println(err)
		}
		logio = io.MultiWriter(os.Stdout,logfile)
		logger.SetOutput(logio)
	})
	go func(){
		for{
			select{
			case strlog := <- l.c:
				if strlog != ""{
					logger.Output(0,strlog)
				}
			}
		}
	}()

}
func NewLog() *stclog{
	return l
}

func getLevel(level string) int8{
	l := strings.ToUpper(level)
	switch l {
	case "TRACE":
		return TRACE
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}


type stclog struct {
	c chan string
	level int8
}

func outstring(a ...interface{}) string {
	n := len(a)
	formats := []string{}
	for i:=0 ;i<n;i++{
		formats = append(formats,"%v")
	}
	format := strings.Join(formats,"")
	res := fmt.Sprintf(format,a...)
	_,file,line,ok := runtime.Caller(2)
	if ok{
		return fmt.Sprintf("%s:%d -> %s",file,line,res)
	}
	return res
}

func outstringformat(format string,a ...interface{}) string {
	_,file,line,ok := runtime.Caller(2)
	if ok{
		arr := []interface{}{}
		arr = append(arr,file,line)
		arr = append(arr,a...)
		return fmt.Sprintf("%s:%d -> "+format,arr...)
	}
	return fmt.Sprintf(format,a...)
}


func rotateLog(handler func()){
	go func() {
		for{
			dur := getNextDateDuration()
			timer := time.NewTimer(dur)
			<-timer.C
			handler()
		}
	}()
}
func getNextDateDuration() time.Duration{
	now := time.Now()
	timeStr := now.Format("2006-01-02")
	dateNow,_ := time.ParseInLocation("2006-01-02",timeStr,time.Local)
	nextDate := dateNow.AddDate(0,0,1)
	diff := nextDate.Sub(now)
	return diff
}



func (l *stclog) Trace(a ...interface{}){
	if(l.level <= TRACE){
		res := outstring(a...)
		if res != ""{
			l.c <- "TRACE "+res
		}
	}
}

func (l *stclog) Tracef(format string ,a ...interface{}) {
	if(l.level <= TRACE){
		res := outstringformat(format,a...)
		if res != ""{
			l.c <- "TRACE "+res
		}
	}
}

func (l *stclog) Debug(a ...interface{}){
	if(l.level <= DEBUG){
		res := outstring(a...)
		if res != ""{
			l.c <- "DEBUG "+res
		}
	}
}

func (l *stclog) Debugf(format string ,a ...interface{}) {
	if(l.level <= DEBUG){
		res := outstringformat(format,a...)
		if res != ""{
			l.c <- "DEBUG "+res
		}
	}
}



func (l *stclog) Info(a ...interface{}){
	if(l.level <= INFO){
		res := outstring(a...)
		if res != ""{
			l.c <- "INFO "+res
		}
	}
}

func (l *stclog) Infof(format string ,a ...interface{}) {
	if(l.level <= INFO){
		res := outstringformat(format,a...)
		if res != ""{
			l.c <- "INFO "+res
		}
	}
}


func (l *stclog) Warn(a ...interface{}){
	if(l.level <= WARN){
		res := outstring(a...)
		if res != ""{
			l.c <- "WARN "+res
		}
	}
}


func (l *stclog) Warnf(format string ,a ...interface{}) {
	if(l.level <= WARN){
		res := outstringformat(format,a...)
		if res != ""{
			l.c <- "WARN "+res
		}
	}
}

func (l *stclog) Error(a ...interface{}){
	if(l.level <= ERROR){
		res := outstring(a...)
		if res != ""{
			l.c <- "ERROR "+res
		}
	}
}

func (l *stclog) Errorf(format string ,a ...interface{}) {
	if(l.level <= ERROR){
		res := outstringformat(format,a...)
		if res != ""{
			l.c <- "ERROR "+res
		}
	}
}

func (l *stclog) Fatal(a ...interface{}){
	res := outstring(a...)
	if res != ""{
		l.c <- "FATAL "+res
	}
	time.Sleep(time.Second*2)
	os.Exit(1)
}

func (l *stclog) Fatalf(format string, a ...interface{}){
	res := outstringformat(format,a...)
	if res != ""{
		l.c <- "FATAL "+res
	}
	time.Sleep(time.Second*2)
	os.Exit(1)
}

