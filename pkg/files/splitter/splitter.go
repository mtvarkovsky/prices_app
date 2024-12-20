package splitter

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"prices/pkg/config"
	"prices/pkg/files"
	"sync"

	"go.uber.org/zap"
)

type (
	FileLines struct {
		File   files.File
		Lines  [][]string
		Parent files.File
	}

	FileQueue interface {
		Put(file files.File) error
		Data() (<-chan files.File, error)
	}

	V1 struct {
		wg         *sync.WaitGroup
		wgInternal *sync.WaitGroup
		config     *config.FileProcessor
		stop       <-chan bool
		fileLines  chan FileLines
		files      FileQueue
		logger     *zap.Logger
	}
)

func NewSplitter(
	wg *sync.WaitGroup,
	logger *zap.Logger,
	config *config.FileProcessor,
	splitFiles FileQueue,
	stop <-chan bool,
) *V1 {
	log := logger.Named("FileSplitter")
	wgInternal := &sync.WaitGroup{}
	s := &V1{
		wg:         wg,
		wgInternal: wgInternal,
		config:     config,
		stop:       stop,
		files:      splitFiles,
		fileLines:  make(chan FileLines, config.FileSplitter.FileLinesQueueSize),
		logger:     log,
	}
	return s
}

func (s *V1) Split() {
	s.logger.Sugar().Infof("start file V1")
	s.wg.Add(1)
	go s.splitFiles()
	for i := 0; i < s.config.FileSplitter.WorkersCount; i++ {
		s.logger.Sugar().Infof("start file V1 worker")
		s.wgInternal.Add(1)
		go s.processSplits()
	}
	<-s.stop
	s.logger.Sugar().Infof("stopping file V1")
	close(s.fileLines)
	s.wgInternal.Wait()
	s.logger.Sugar().Infof("stop file V1")
	s.wg.Done()
}

func (s *V1) splitFiles() {
	fls, err := s.files.Data()
	if err != nil {
		s.logger.Sugar().Infof("can't get file queue data: (%s)", err.Error())
		return
	}
	for file := range fls {
		s.wgInternal.Add(1)
		go s.splitFile(file)
	}
}

func (s *V1) splitFile(file files.File) {
	s.logger.Sugar().Infof("try to split file=%s", file)
	defer s.wgInternal.Done()
	f, err := os.Open(file.Path)
	if err != nil {
		s.logger.Sugar().Errorf("can't open file=%s: (%s)", file, err.Error())
		return
	}
	reader := csv.NewReader(f)
	var lines [][]string
	counter := 0
	for {
		line, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				if len(lines) > 0 {
					s.pushFileLines(file, lines, counter-len(lines), counter)
				}
				s.logger.Sugar().Infof("done reading file=%s", file)
				_ = f.Close()
				break
			}
			s.logger.Sugar().Errorf("can't read file=%s data: (%s)", file, err.Error())
			return
		}
		counter += 1
		lines = append(lines, line)
		if len(lines)%s.config.FileSplitter.SplitByLines == 0 {
			s.pushFileLines(file, lines, counter-len(lines), counter)
			lines = nil
		}
	}
	s.logger.Sugar().Infof("done splitting file=%s", file)
}

func (s *V1) pushFileLines(file files.File, lines [][]string, start int, end int) {
	fileLines := FileLines{
		File: files.File{
			Path: fmt.Sprintf("%s/%d_%d_%s", s.config.FilesDir, start, end, file.Name),
		},
		Lines:  lines,
		Parent: file,
	}
	s.fileLines <- fileLines
}

func (s *V1) processSplits() {
	defer func() {
		s.logger.Sugar().Infof("stop file V1 worker")
		s.wgInternal.Done()
	}()
	for fl := range s.fileLines {
		f, err := os.Create(fl.File.Path)
		if err != nil {
			s.logger.Sugar().Errorf("can't to create file=%s: (%s)", fl.File.Path, err.Error())
			s.wgInternal.Done()
			return
		}
		writer := csv.NewWriter(f)
		err = writer.WriteAll(fl.Lines)
		if err != nil {
			s.logger.Sugar().Errorf("can't write file=%s: (%s)", fl.File.Path, err.Error())
		}
		f.Close()
	}
}
