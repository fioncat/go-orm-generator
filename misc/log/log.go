package log

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fioncat/go-gendb/misc/term"
)

var (
	logEnable bool
	logPath   string

	warnLogFile bool
)

const (
	logFmt  = "[%s] %s | %s"
	timeFmt = "2006-01-02 15:04:05"
)

// IsEnable returns whether the log is enabled
func IsEnable() bool {
	return logEnable
}

func appendLogFile(msg string) error {
	file, err := os.OpenFile(logPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(msg + "\n")
	return err
}

// Init initialization log.
// By default, the log will be output to standard output,
// if path is not empty, it will be output to the file
// specified by path.
func Init(enable bool, path string) {
	logEnable = enable
	logPath = path
}

// returns level string(with color)
func getLevel(level string) string {
	if logPath != "" {
		return level
	}
	switch level {
	case "INF":
		return term.Info(level)
	case "ERR":
		return term.Red(level)
	}
	return level
}

func writeLog(level, msg string) {
	now := time.Now().Format(timeFmt)
	level = getLevel(level)
	log := fmt.Sprintf(logFmt, level, now, msg)
	if logPath != "" {
		err := appendLogFile(log)
		if err != nil && !warnLogFile {
			fmt.Printf("write log failed: %v\n", err)
			warnLogFile = true
		}
		return
	}
	fmt.Println(log)
}

func joinI(vals ...interface{}) string {
	strs := make([]string, len(vals))
	for i, s := range vals {
		strs[i] = fmt.Sprint(s)
	}
	return strings.Join(strs, " ")
}

// Info output INFO level log
func Info(ss ...interface{}) {
	if !logEnable {
		return
	}
	msg := joinI(ss...)
	writeLog("INF", msg)
}

// Infof output INFO level log
func Infof(layer string, vs ...interface{}) {
	if !logEnable {
		return
	}
	msg := fmt.Sprintf(layer, vs...)
	writeLog("INF", msg)
}

// Error output ERR level log
func Error(ss ...interface{}) {
	if !logEnable {
		return
	}
	msg := joinI(ss...)
	writeLog("ERR", msg)
}

// Errorf output ERR level log
func Errorf(layer string, vs ...interface{}) {
	if !logEnable {
		return
	}
	msg := fmt.Sprintf(layer, vs...)
	writeLog("ERR", msg)
}

// Fatal output log and exit program with 1.
func Fatal(layer string, vs ...interface{}) {
	msg := fmt.Sprintf(layer, vs...)
	fmt.Printf("%s: %s\n", term.Red("fatal"), msg)
	os.Exit(1)
}
