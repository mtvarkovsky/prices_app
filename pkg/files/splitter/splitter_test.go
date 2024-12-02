package splitter

import (
	"encoding/csv"
	"fmt"
	"os"
	"prices/pkg/config"
	"prices/pkg/files"
	"prices/pkg/testutils"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func newTestSplitter(t *testing.T) (*V1, chan bool) {
	dir, err := os.MkdirTemp("", "")
	assert.NoError(t, err)

	splitFiles := files.NewFileQueueInMem(1)

	wg := &sync.WaitGroup{}
	log := zap.NewNop()

	cfg := &config.FileProcessor{
		FilesDir:         dir,
		MaxFileSizeBytes: 6000,
		FileSplitter: config.FileSplitter{
			SplitByLines:       50,
			WorkersCount:       2,
			FileLinesQueueSize: 2,
		},
	}

	stop := make(chan bool)

	splttr := NewSplitter(wg, log, cfg, splitFiles, stop)

	return splttr, stop
}

func TestSplitter_PushFileLines(t *testing.T) {
	splttr, _ := newTestSplitter(t)
	dir := splttr.config.FilesDir

	file := files.File{
		Path: "test",
	}
	lines := [][]string{
		{"id_1", "price_1", "expirationDate_1"},
		{"id_1", "price_2", "expirationDate_2"},
	}

	expectedLines := FileLines{
		File: files.File{
			Path: fmt.Sprintf("%s/%d_%d_%s", dir, 0, 1, file.Name),
		},
		Lines:  lines,
		Parent: file,
	}

	go splttr.pushFileLines(file, lines, 0, 1)

	resLines := <-splttr.fileLines

	assert.Equal(t, expectedLines, resLines)
}

func TestSplitter_SplitFile(t *testing.T) {
	splttr, _ := newTestSplitter(t)
	dir := splttr.config.FilesDir
	testutils.GenerateTestData(100, dir)

	entries, err := os.ReadDir(dir)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)

	entry := entries[0]

	file := files.File{
		Path: fmt.Sprintf("%s/%s", dir, entry.Name()),
		Name: entry.Name(),
	}

	expectedFile1 := files.File{
		Path: fmt.Sprintf("%s/%d_%d_%s", dir, 0, 50, file.Name),
	}
	var expectedFile1Lines [][]string
	expectedFile2 := files.File{
		Path: fmt.Sprintf("%s/%d_%d_%s", dir, 50, 100, file.Name),
	}
	var expectedFile2Lines [][]string

	f, err := os.Open(file.Path)
	assert.NoError(t, err)

	reader := csv.NewReader(f)
	lines, err := reader.ReadAll()
	assert.NoError(t, err)

	f.Close()

	expectedFile1Lines = lines[0:50]
	expectedFile2Lines = lines[50:]

	splttr.wgInternal.Add(1)
	go splttr.splitFile(file)

	lines1 := <-splttr.fileLines
	assert.Equal(t, expectedFile1, lines1.File)
	assert.Equal(t, file, lines1.Parent)
	assert.Equal(t, expectedFile1Lines, lines1.Lines)

	lines2 := <-splttr.fileLines
	assert.Equal(t, expectedFile2, lines2.File)
	assert.Equal(t, file, lines2.Parent)
	assert.Equal(t, expectedFile2Lines, lines2.Lines)
}

func TestSplitter_ProcessSplits(t *testing.T) {
	splttr, _ := newTestSplitter(t)
	dir := splttr.config.FilesDir

	file := files.File{
		Path: fmt.Sprintf("%s/%s", dir, "test.csv"),
		Name: "test.csv",
	}

	lines := FileLines{
		File: files.File{
			Path: fmt.Sprintf("%s/%d_%d_%s", dir, 0, 1, file.Name),
		},
		Lines: [][]string{
			{"id_1", "price_1", "expirationDate_1"},
			{"id_1", "price_2", "expirationDate_2"},
		},
		Parent: file,
	}

	go splttr.processSplits()
	splttr.fileLines <- lines

	for {
		entries, err := os.ReadDir(dir)
		assert.NoError(t, err)
		if len(entries) == 1 {
			break
		}
	}

	f, err := os.Open(lines.File.Path)
	assert.NoError(t, err)
	reader := csv.NewReader(f)

	resLines, err := reader.ReadAll()
	assert.NoError(t, err)

	assert.Equal(t, lines.Lines, resLines)
}

func TestSplitter_Split(t *testing.T) {
	splttr, stop := newTestSplitter(t)
	dir := splttr.config.FilesDir

	filesQ := splttr.files.(*files.FileQueueInMem)

	testutils.GenerateTestData(100, dir)

	entries, err := os.ReadDir(dir)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)

	entry := entries[0]

	file := files.File{
		Path: fmt.Sprintf("%s/%s", dir, entry.Name()),
		Name: entry.Name(),
	}

	expectedFile1 := files.File{
		Path: fmt.Sprintf("%s/%d_%d_%s", dir, 0, 50, file.Name),
	}
	var expectedFile1Lines [][]string
	expectedFile2 := files.File{
		Path: fmt.Sprintf("%s/%d_%d_%s", dir, 50, 100, file.Name),
	}
	var expectedFile2Lines [][]string

	f, err := os.Open(file.Path)
	assert.NoError(t, err)

	reader := csv.NewReader(f)
	lines, err := reader.ReadAll()
	assert.NoError(t, err)

	f.Close()

	expectedFile1Lines = lines[0:50]
	expectedFile2Lines = lines[50:]

	go splttr.Split()

	err = filesQ.Put(file)
	assert.NoError(t, err)

	for {
		entries, err := os.ReadDir(dir)
		assert.NoError(t, err)
		if len(entries) == 3 {
			break
		}
	}

	stop <- true
	err = filesQ.Close()
	assert.NoError(t, err)

	_, open := <-splttr.fileLines
	assert.False(t, open)

	f1, err := os.Open(expectedFile1.Path)
	assert.NoError(t, err)
	reader = csv.NewReader(f1)
	f1Lines, err := reader.ReadAll()
	assert.NoError(t, err)
	assert.Equal(t, expectedFile1Lines, f1Lines)
	f1.Close()

	f2, err := os.Open(expectedFile2.Path)
	assert.NoError(t, err)
	reader = csv.NewReader(f2)
	f2Lines, err := reader.ReadAll()
	assert.NoError(t, err)
	assert.Equal(t, expectedFile2Lines, f2Lines)
	f2.Close()

	_, err = os.Stat(file.Path)
	assert.NoError(t, err)
}
