package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/crowdsecurity/crowdsec/pkg/types"
	log "github.com/sirupsen/logrus"

	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v2"
)

type bouncerConfig struct {
	BinPath                string        `yaml:"bin_path"` // path to binary
	PidDir                 string        `yaml:"piddir"`
	UpdateFrequency        string        `yaml:"update_frequency"`
	Daemon                 bool          `yaml:"daemonize"`
	LogMode                string        `yaml:"log_mode"`
	LogDir                 string        `yaml:"log_dir"`
	LogLevel               log.Level     `yaml:"log_level"`
	CompressLogs           *bool         `yaml:"compress_logs,omitempty"`
	LogMaxSize             int           `yaml:"log_max_size,omitempty"`
	LogMaxFiles            int           `yaml:"log_max_files,omitempty"`
	LogMaxAge              int           `yaml:"log_max_age,omitempty"`
	CacheRetentionDuration time.Duration `yaml:"cache_retention_duration"`
}

func NewConfig(configPath string) (*bouncerConfig, error) {
	var LogOutput *lumberjack.Logger //io.Writer

	config := &bouncerConfig{}

	configBuff, err := ioutil.ReadFile(configPath)
	if err != nil {
		return &bouncerConfig{}, fmt.Errorf("failed to read %s : %v", configPath, err)
	}

	err = yaml.Unmarshal(configBuff, &config)
	if err != nil {
		return &bouncerConfig{}, fmt.Errorf("failed to unmarshal %s : %v", configPath, err)
	}

	if config.BinPath == "" {
		return &bouncerConfig{}, fmt.Errorf("bin_path is not set")
	}
	if config.LogMode == "" {
		return &bouncerConfig{}, fmt.Errorf("log_mode is not net")
	}

	_, err = os.Stat(config.BinPath)
	if os.IsNotExist(err) {
		return config, fmt.Errorf("binary '%s' doesn't exist", config.BinPath)
	}

	/*Configure logging*/
	if err = types.SetDefaultLoggerConfig(config.LogMode, config.LogDir, config.LogLevel, config.LogMaxSize, config.LogMaxFiles, config.LogMaxAge, config.CompressLogs, false); err != nil {
		log.Fatal(err.Error())
	}
	if config.LogMode == "file" {
		if config.LogDir == "" {
			config.LogDir = "/var/log/"
		}
		LogOutput = &lumberjack.Logger{
			Filename:   config.LogDir + "/crowdsec-custom-bouncer.log",
			MaxSize:    500, //megabytes
			MaxBackups: 3,
			MaxAge:     28,   //days
			Compress:   true, //disabled by default
		}
		log.SetOutput(LogOutput)
		log.SetFormatter(&log.TextFormatter{TimestampFormat: "02-01-2006 15:04:05", FullTimestamp: true})
	} else if config.LogMode != "stdout" {
		return &bouncerConfig{}, fmt.Errorf("log mode '%s' unknown, expecting 'file' or 'stdout'", config.LogMode)
	}

	if config.CacheRetentionDuration == 0 {
		log.Infof("cache_retention_duration defaults to 10 seconds")
		config.CacheRetentionDuration = time.Duration(10 * time.Second)
	}

	return config, nil
}
