package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type (
	Config[C any] interface {
		LoadConfig(name string) (C, error)
	}

	FileProcessor struct {
		FilesDir                string       `mapstructure:"FILES_DIRECTORY"`
		FilesProcessedDir       string       `mapstructure:"FILES_PROCESSED_DIRECTORY"`
		FilesErrorsDir          string       `mapstructure:"FILES_ERRORS_DIRECTORY"`
		FilesQueueSize          int          `mapstructure:"FILES_QUEUE_SIZE"`
		FilesProcessedQueueSize int          `mapstructure:"FILES_PROCESSED_QUEUE_SIZE"`
		FilesErrorQueueSize     int          `mapstructure:"FILES_ERROR_QUEUE_SIZE"`
		FilesSplitQueueSize     int          `mapstructure:"FILES_SPLIT_QUEUE_SIZE"`
		MaxFileSizeBytes        int64        `mapstructure:"MAX_FILE_SIZE_BYTES"`
		DataBatchSize           int          `mapstructure:"DATA_BATCH_SIZE"`
		DataBatchQueueSize      int          `mapstructure:"DATA_BATCH_QUEUE_SIZE"`
		WorkersCount            int          `mapstructure:"WORKERS_COUNT"`
		ImportByLines           bool         `mapstructure:"IMPORT_BY_LINES"`
		FileScanner             FileScanner  `mapstructure:"FILE_SCANNER"`
		FileSplitter            FileSplitter `mapstructure:"FILE_SPLITTER"`
		Storage                 Storage      `mapstructure:"STORAGE"`
	}

	FileScanner struct {
		CheckEveryDuration time.Duration `mapstructure:"CHECK_EVERY_DURATION"`
	}

	FileSplitter struct {
		WorkersCount       int `mapstructure:"WORKERS_COUNT"`
		FileLinesQueueSize int `mapstructure:"LINES_QUEUE_SIZE"`
		SplitByLines       int `mapstructure:"SPLIT_BY_LINES"`
	}

	APIServer struct {
		Port    int     `mapstructure:"PORT"`
		Storage Storage `mapstructure:"STORAGE"`
	}

	Storage struct {
		Type           string `mapstructure:"TYPE"`
		DSN            string `mapstructure:"DSN"`
		MaxConnections int    `mapstructure:"MAX_CONNECTIONS"`
	}
)

func (cfg *FileProcessor) LoadConfig(name string) (*FileProcessor, error) {
	viper.AddConfigPath("./configs")
	viper.SetConfigName(name)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("can't read config for FileProcessor from=%s: %w", name, err)
	}

	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshall config for FileProcessor from=%s: %w", name, err)
	}

	return cfg, nil
}

func (cfg *APIServer) LoadConfig(name string) (*APIServer, error) {
	viper.AddConfigPath("./configs")
	viper.SetConfigName(name)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("can't read config for APIServer from=%s: %w", name, err)
	}

	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshall config for APIServer from=%s: %w", name, err)
	}

	return cfg, nil
}
