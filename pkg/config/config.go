package config

import (
	"encoding/json"
	"gopkg.in/yaml.v2"
	"os"
)

type Config struct {
	// NoClientAuth is true if clients are allowed to connect without authenticating.
	NoClientAuth  bool     `json:"no_client_auth" yaml:"no_client_auth"`
	PasswordAuth  *Auth    `json:"password_auth" yaml:"password_auth"`
	PublicKeyAuth *Auth    `json:"public_key_auth" yaml:"public_key_auth"`
	HostKeys      []string `json:"host_keys" yaml:"host_keys"`

	// MaxAuthTries specifies the maximum number of authentication attempts
	// permitted per connection. If set to a negative number, the number of
	// attempts are unlimited. If set to zero, the number of attempts are limited
	// to 6.
	MaxAuthTries int `json:"max_auth_tries" yaml:"max_auth_tries"`
	// ServerVersion is the version identification string to announce in
	// the public handshake.
	// If empty, a reasonable default is used.
	// Note that RFC 4253 section 4.2 requires that this string start with
	// "SSH-2.0-".
	ServerVersion string `json:"server_version" yaml:"server_version"`

	// BannerCallback, if present, is called and the return string is sent to
	// the client after key exchange completed but before authentication.
	Banner string `json:"banner" yaml:"banner"`

	Addr string     `json:"addr" yaml:"addr"`
	Log  *LogConfig `json:"log" yaml:"log"`
}
type Auth struct {
	Enable bool `json:"enable" yaml:"enable"`
	Accept bool `json:"accept" yaml:"accept"`
}
type LogConfig struct {
	IsDebug        bool          `json:"debug" yaml:"debug"`
	IsFileOut      bool          `json:"file_out" yaml:"file_out"`
	FileName       string        `json:"file_name" yaml:"file_name"`
	FileOnlySSHLog bool          `json:"file_only_ssh_log" yaml:"file_only_ssh_log"`
	IsJson         bool          `json:"json" yaml:"json"`
	SeparatePort   bool          `json:"separate_port" yaml:"separate_port"`
	RotateConf     *RotateConfig `json:"rotate" yaml:"rotate"`
}
type RotateConfig struct {
	Enable bool `yaml:"enable"`
	// Filename is the file to write logs to.  Backup log files will be retained
	// in the same directory.  It uses <processname>-lumberjack.log in
	// os.TempDir() if empty.
	Filename string `json:"filename" yaml:"filename"`
	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes.
	MaxSize int `json:"max_size" yaml:"max_size"`
	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age.
	MaxAge int `json:"max_age" yaml:"max_age"`
	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxAge may still cause them to get
	// deleted.)
	MaxBackups int `json:"max_back_ups" yaml:"max_back_ups"`
	// LocalTime determines if the time used for formatting the timestamps in
	// backup files is the computer's local time.  The default is to use UTC
	// time.
	LocalTime bool `json:"localtime" yaml:"localtime"`
	// Compress determines if the rotated log files should be compressed
	// using gzip. The default is not to perform compression.
	Compress bool `json:"compress" yaml:"compress"`
}

func (c *Config) String() string {
	marshal, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return string(marshal)
}

func DefaultConfig() *Config {
	return &Config{
		NoClientAuth:  false,
		PasswordAuth:  &Auth{Enable: true, Accept: false},
		PublicKeyAuth: &Auth{Enable: true, Accept: false},
		HostKeys:      []string{"host.key"},
		MaxAuthTries:  6,
		ServerVersion: "SSH-2.0-OpenSSH_7.4",
		Banner:        "WARNING: YOU ARE BEING MONITORED!",
		Addr:          "0.0.0.0:2023",
		Log: &LogConfig{
			IsDebug:        false,
			IsFileOut:      false,
			FileName:       "./logs/ssh-honeypot.log",
			FileOnlySSHLog: false,
			IsJson:         false,
			SeparatePort:   false,
			RotateConf: &RotateConfig{
				Enable:     false,
				Filename:   "./logs/ssh-honeypot-rotate.log",
				MaxSize:    100,
				MaxAge:     0,
				MaxBackups: 0,
				LocalTime:  true,
				Compress:   false,
			},
		},
	}
}
func GenConf(config *Config, path string) error {
	bytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, bytes, 0644)
	if err != nil {
		return err
	}
	return err
}

func ParseConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	conf := DefaultConfig()
	err = yaml.Unmarshal(file, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
