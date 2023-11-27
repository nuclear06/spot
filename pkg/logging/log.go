package logging

import (
	"fmt"
	"github.com/sirupsen/logrus"
	. "github.com/ssh-honeypot/pkg/config"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var sepPort = false
var log = New()
var sshLog = New()

type Event int

const (
	_ Event = iota
	SysInit
	SysRunning
	NoAuth
	PasswordAuth
	PublicKeyAuth
	SshEnv
	SshExec
	SshShell
	SshPtyReq
	SshOther
)

func (e Event) E() string {
	return [...]string{"", "sys_init", "sys_running", "no_auth", "password_auth", "public_key_auth", "ssh_env", "ssh_exec", "ssh_shell", "ssh_pty_req", "ssh_other"}[e]
}

type Logger struct {
	*logrus.Logger
}

func New() *Logger {
	return &Logger{logrus.New()}
}

func InitLog(conf Config) *Logger {
	LogConf := conf.Log
	rotateConf := conf.Log.RotateConf
	sepPort = LogConf.SeparatePort
	if LogConf.IsDebug {
		SetLevel(logrus.DebugLevel)
	}
	if LogConf.IsJson {
		log.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339})
		sshLog.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339})
	} else {
		log.SetFormatter(&logrus.TextFormatter{TimestampFormat: time.RFC3339})
		sshLog.SetFormatter(&logrus.TextFormatter{TimestampFormat: time.RFC3339})
	}
	if LogConf.IsFileOut {
		// rotate log
		if rotateConf.Enable {
			lumberjackLogger := &lumberjack.Logger{
				Filename:   rotateConf.Filename,
				MaxSize:    rotateConf.MaxSize,
				MaxBackups: rotateConf.MaxBackups,
				MaxAge:     rotateConf.MaxAge,
				LocalTime:  rotateConf.LocalTime,
				Compress:   rotateConf.Compress,
			}
			mWriter := io.MultiWriter(os.Stdout, lumberjackLogger)
			if LogConf.FileOnlySSHLog {
				sshLog.SetOutput(mWriter)
			} else {
				sshLog.SetOutput(mWriter)
				log.SetOutput(mWriter)
			}
		} else {
			// no rotate, direct to file
			dir := filepath.Dir(LogConf.FileName)
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				LogWarn(SysInit, fmt.Sprintf("failed to create log dir, err: [%v]", err))
			}

			file, err := os.OpenFile(LogConf.FileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			mWriter := io.MultiWriter(os.Stdout, file)
			if err == nil {
				if LogConf.FileOnlySSHLog {
					sshLog.SetOutput(mWriter)
				} else {
					sshLog.SetOutput(mWriter)
					log.SetOutput(mWriter)
				}
			} else {
				LogWarn(SysInit, "failed to open log file, using default stdOut")
			}
		}
	}

	LogDebug(SysInit, "Logger init finished")
	SSHDebug(SysInit, "", "SSHLogger init finished", nil)
	return log
}

func SetLevel(level logrus.Level) {
	log.SetLevel(level)
	sshLog.SetLevel(level)
}

func LogFields(level logrus.Level, event Event, msg string) {
	l := log.WithFields(logrus.Fields{
		"type":  "sys",
		"event": event.E(),
	})
	l.Log(level, msg)
}

func SSHFields(level logrus.Level, event Event, addr, msg string, fs logrus.Fields) {
	e := log.WithFields(logrus.Fields{
		"type":  "ssh",
		"event": event.E(),
	})
	e = processAddr(e, addr)
	e.WithFields(fs).Log(level, msg)
}

func processAddr(e *logrus.Entry, addr string) *logrus.Entry {
	if addr == "" {
		return e
	}
	if sepPort {
		s := strings.Split(addr, ":")
		if len(s) != 2 {
			LogFields(logrus.WarnLevel, SysRunning, fmt.Sprintf("split address error, addr: [%s]", addr))
			goto NoPort
		}
		ip, port := s[0], s[1]
		return e.WithFields(logrus.Fields{
			"ip":   ip,
			"port": port,
		})
	}
NoPort:
	return e.WithField("ip", addr)
}
func LogDebug(e Event, msg string) {
	LogFields(logrus.DebugLevel, e, msg)
}
func LogInfo(e Event, msg string) {
	LogFields(logrus.InfoLevel, e, msg)
}
func LogWarn(e Event, msg string) {
	LogFields(logrus.WarnLevel, e, msg)
}
func LogError(e Event, msg string) {
	LogFields(logrus.ErrorLevel, e, msg)
}
func LogFatal(e Event, msg string) {
	LogFields(logrus.FatalLevel, e, msg)
}
func SSHDebug(e Event, addr, msg string, fs logrus.Fields) {
	SSHFields(logrus.DebugLevel, e, addr, msg, fs)
}
func SSHInfo(e Event, addr, msg string, fs logrus.Fields) {
	SSHFields(logrus.InfoLevel, e, addr, msg, fs)
}
func SSHWarn(e Event, addr, msg string, fs logrus.Fields) {
	SSHFields(logrus.WarnLevel, e, addr, msg, fs)
}
func SSHError(e Event, addr, msg string, fs logrus.Fields) {
	SSHFields(logrus.ErrorLevel, e, addr, msg, fs)
}
