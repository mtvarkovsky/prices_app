package files

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"prices/pkg/config"
	"sync"
	"time"
)

const (
	CSV = ".csv"
)

type (
	Scanner interface {
		Scan()
	}

	scanner struct {
		wg         *sync.WaitGroup
		config     *config.FileProcessor
		stop       <-chan bool
		files      FileQueue
		splitFiles FileQueue
		cache      FileCache
		logger     *zap.Logger
	}
)

func NewScanner(wg *sync.WaitGroup, logger *zap.Logger, config *config.FileProcessor, files FileQueue, splitFiles FileQueue, cache FileCache, stop <-chan bool) Scanner {
	log := logger.Named("FileScanner")
	s := &scanner{
		wg:         wg,
		config:     config,
		stop:       stop,
		cache:      cache,
		files:      files,
		splitFiles: splitFiles,
		logger:     log,
	}
	return s
}

func (s *scanner) Scan() {
	s.logger.Sugar().Infof("try to start scanning files in directory=%s", s.config.FilesDir)
	if _, err := os.Stat(s.config.FilesDir); os.IsNotExist(err) {
		s.logger.Sugar().Errorf("can't open directory=%s: (%s)", s.config.FilesDir, err.Error())
	}
	s.wg.Add(1)
	ticker := time.Tick(s.config.FileScanner.CheckEveryDuration)
	s.scanDir()
	for {
		select {
		case <-ticker:
			s.logger.Sugar().Infof("rescan directory=%s", s.config.FilesDir)
			s.scanDir()
		case <-s.stop:
			s.logger.Sugar().Infof("stop scanning files in directory=%s", s.config.FilesDir)
			s.wg.Done()
			return
		}
	}
}

func (s *scanner) scanDir() {
	dir, err := os.ReadDir(s.config.FilesDir)
	if err != nil {
		s.logger.Sugar().Errorf("can't open directory=%s: (%s)", s.config.FilesDir, err.Error())
		return
	}
	for _, entry := range dir {
		if s.valid(entry) {
			s.add(entry)
		}
	}
}

func (s *scanner) valid(entry os.DirEntry) bool {
	if entry.IsDir() {
		return false
	}

	extension := filepath.Ext(s.getPath(entry))
	if extension != CSV {
		return false
	}

	return true
}

func (s *scanner) getPath(entry os.DirEntry) string {
	return filepath.Join(s.config.FilesDir, entry.Name())
}

func (s *scanner) add(entry os.DirEntry) {
	path := s.getPath(entry)
	if _, ok, err := s.cache.Get(path); ok {
		return
	} else if err != nil {
		s.logger.Sugar().Errorf("can't get file=%s from cache: (%s)", path, err.Error())
		return
	}

	s.logger.Sugar().Infof("add entry=%s to files queue", path)

	now := time.Now()
	newName := fmt.Sprintf("%d_%s", now.UnixNano(), entry.Name())
	newPath := fmt.Sprintf("%s/%s", s.config.FilesDir, newName)
	if err := os.Rename(path, newPath); err != nil {
		s.logger.Sugar().Errorf("can't reanme entry=%s to=%s: (%s)", path, newPath, err.Error())
		return
	}
	newFile := File{Path: newPath, Name: newName}

	newFileInfo, err := os.Stat(newFile.Path)
	if err != nil {
		s.logger.Sugar().Errorf("can't get file=%s info: (%s)", newFile, err.Error())
		return
	}
	if newFileInfo.Size() >= s.config.MaxFileSizeBytes {
		if err := s.cache.Put(newFile); err != nil {
			s.logger.Sugar().Errorf("can't add file=%s to cache: (%s)", newFile, err.Error())
			return
		}
		if err := s.splitFiles.Put(newFile); err != nil {
			s.logger.Sugar().Errorf("can't add file=%s to splitFile queue: (%s)", newFile, err.Error())
			return
		}
		return
	}

	if err := s.cache.Put(newFile); err != nil {
		s.logger.Sugar().Errorf("can't add entry=%s to cache: (%s)", newFile, err.Error())
		return
	}
	if err := s.files.Put(newFile); err != nil {
		s.logger.Sugar().Errorf("can't add entry=%s to files queue: (%s)", newFile, err.Error())
		return
	}
}
