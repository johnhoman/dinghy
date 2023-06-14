package logging

import (
	"fmt"
	"log"
	"os"
	"sync"
)

var (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	White  = "\033[97m"
)

type Logger struct {
	log *log.Logger
	mu  *sync.Mutex
	pre string
}

func New() *Logger {
	return &Logger{
		log: log.New(os.Stderr, "", 0),
		mu:  &sync.Mutex{},
	}
}

func (log *Logger) SetPrefix(pre string) {
	log.mu.Lock()
	log.pre = pre + " "
	log.mu.Unlock()
}

func (log *Logger) Debug(format string, args ...any) {
	log.log.Printf(log.pre+Gray+"DEBUG: %s"+Reset, fmt.Sprintf(format, args...))
}

func (log *Logger) Ok(format string, args ...any) {
	log.log.Printf(log.pre+Green+"OK:   %s %s", Reset, fmt.Sprintf(format, args...))
}

func (log *Logger) Error(format string, args ...any) {
	log.log.Printf(log.pre+Red+"ERROR:%s %s", Reset, fmt.Sprintf(format, args...))
}

func (log *Logger) Fatal(format string, args ...any) {
	log.Error(format, args...)
	os.Exit(1)
}

func UseRed(s string) string {
	return Red + s + Reset
}

func UseGray(s string) string {
	return Gray + s + Reset
}

func UseGreen(s string) string {
	return Green + s + Reset
}
