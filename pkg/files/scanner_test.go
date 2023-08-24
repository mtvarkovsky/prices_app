package files

import (
	"os"
	"prices/pkg/config"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func newTestScanner(t *testing.T) (*scanner, chan bool) {
	dir, err := os.MkdirTemp("", "")
	assert.NoError(t, err)

	files := NewFileQueueInMem(0)
	splitFiles := NewFileQueueInMem(1)
	cache := NewFileCacheInMem()

	wg := &sync.WaitGroup{}
	log := zap.NewNop()

	cfg := &config.FileProcessor{
		FilesDir:         dir,
		MaxFileSizeBytes: 1 << 20,
		FileScanner: config.FileScanner{
			CheckEveryDuration: time.Microsecond,
		},
	}

	stop := make(chan bool)

	scnnr := NewScanner(wg, log, cfg, files, splitFiles, cache, stop)
	s, ok := scnnr.(*scanner)
	assert.True(t, ok)

	return s, stop
}

func TestScanner_GetPath(t *testing.T) {
	scnnr, _ := newTestScanner(t)
	dir := scnnr.config.FilesDir
	file, err := os.CreateTemp(dir, "*.csv")
	assert.NoError(t, err)

	entries, err := os.ReadDir(dir)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)

	expected := file.Name()
	path := scnnr.getPath(entries[0])
	assert.Equal(t, expected, path)
}

func TestScanner_Valid_CSV(t *testing.T) {
	scnnr, _ := newTestScanner(t)
	dir := scnnr.config.FilesDir
	_, err := os.CreateTemp(dir, "*.csv")
	assert.NoError(t, err)

	entries, err := os.ReadDir(dir)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)

	assert.True(t, scnnr.valid(entries[0]))
}

func TestScanner_Valid_DIR(t *testing.T) {
	scnnr, _ := newTestScanner(t)
	dir := scnnr.config.FilesDir
	_, err := os.MkdirTemp(dir, "")
	assert.NoError(t, err)

	entries, err := os.ReadDir(dir)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)

	assert.False(t, scnnr.valid(entries[0]))
}

func TestScanner_Valid_TXT(t *testing.T) {
	scnnr, _ := newTestScanner(t)
	dir := scnnr.config.FilesDir
	_, err := os.CreateTemp(dir, "*.txt")
	assert.NoError(t, err)

	entries, err := os.ReadDir(dir)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)

	assert.False(t, scnnr.valid(entries[0]))
}

func TestScanner_Add_NewFile(t *testing.T) {
	scnnr, _ := newTestScanner(t)
	dir := scnnr.config.FilesDir
	file, err := os.CreateTemp(dir, "*.csv")
	assert.NoError(t, err)

	files := scnnr.files
	cache := scnnr.cache

	entries, err := os.ReadDir(dir)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)

	_, ok, _ := cache.Get(file.Name())
	assert.False(t, ok)

	testWg := &sync.WaitGroup{}

	testWg.Add(1)
	go func() {
		scnnr.add(entries[0])
		testWg.Done()
	}()

	var newFile File

	testWg.Add(1)
	go func() {
		newFile, err = files.Get()
		assert.NoError(t, err)
		testWg.Done()
	}()

	testWg.Wait()

	_, ok, _ = cache.Get(newFile.Path)
	assert.True(t, ok)

	_, err = os.Stat(file.Name())
	assert.Error(t, err)

	_, err = os.Stat(newFile.Path)
	assert.NoError(t, err)
}

func TestScanner_Add_NewFile_PushToSplitQueue(t *testing.T) {
	scnnr, _ := newTestScanner(t)
	dir := scnnr.config.FilesDir
	file, err := os.CreateTemp(dir, "*.csv")
	assert.NoError(t, err)
	err = file.Truncate(2 << 20)
	assert.NoError(t, err)

	files := scnnr.files
	splitFiles := scnnr.splitFiles
	cache := scnnr.cache

	entries, err := os.ReadDir(dir)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)

	_, ok, _ := cache.Get(file.Name())
	assert.False(t, ok)

	testWg := &sync.WaitGroup{}

	testWg.Add(1)
	go func() {
		scnnr.add(entries[0])
		testWg.Done()
	}()

	testWg.Wait()

	empty := files.Empty()
	assert.True(t, empty)

	newFile, err := splitFiles.Get()
	assert.NoError(t, err)

	_, ok, _ = cache.Get(newFile.Path)
	assert.True(t, ok)

	_, err = os.Stat(file.Name())
	assert.Error(t, err)

	_, err = os.Stat(newFile.Path)
	assert.NoError(t, err)
}

func TestScanner_Add_FileAlreadyInCache(t *testing.T) {
	scnnr, _ := newTestScanner(t)
	dir := scnnr.config.FilesDir
	file, err := os.CreateTemp(dir, "*.csv")
	assert.NoError(t, err)

	files := scnnr.files
	cache := scnnr.cache

	entries, err := os.ReadDir(dir)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)

	err = cache.Put(File{Path: file.Name()})
	assert.NoError(t, err)

	testWg := &sync.WaitGroup{}

	testWg.Add(1)
	go func() {
		scnnr.add(entries[0])
		testWg.Done()
	}()

	testWg.Wait()

	assert.True(t, files.Empty())
	assert.Equal(t, cache.Len(), 1)
}

func TestScanner_Scan(t *testing.T) {
	scnnr, _ := newTestScanner(t)
	dir := scnnr.config.FilesDir

	files := scnnr.files
	cache := scnnr.cache

	go scnnr.Scan()

	testWg := &sync.WaitGroup{}

	numFiles := 10
	testWg.Add(2 * numFiles)

	counter := 0

	go func() {
		data, err := files.Data()
		assert.NoError(t, err)
		for range data {
			counter += 1
			testWg.Done()
		}
	}()

	go func() {
		for i := 0; i < numFiles; i++ {
			_, err := os.CreateTemp(dir, "*.csv")
			assert.NoError(t, err)
			time.Sleep(time.Microsecond)
			testWg.Done()
		}
	}()

	testWg.Wait()

	assert.Equal(t, numFiles, counter)
	assert.Equal(t, cache.Len(), numFiles)
}

func TestScanner_Scan_Stop(t *testing.T) {
	scnnr, stop := newTestScanner(t)
	dir := scnnr.config.FilesDir

	wg := scnnr.wg
	files := scnnr.files
	cache := scnnr.cache

	go scnnr.Scan()

	testWg := &sync.WaitGroup{}

	numFiles := 10
	expectedFilesAfterStop := numFiles / 2
	testWg.Add(numFiles / 2)

	counter := 0

	go func() {
		data, err := files.Data()
		assert.NoError(t, err)
		for range data {
			counter += 1
			testWg.Done()
			if counter == expectedFilesAfterStop {
				stop <- true
				break
			}
		}
	}()

	go func() {
		for i := 0; i < numFiles; i++ {
			_, err := os.CreateTemp(dir, "*.csv")
			assert.NoError(t, err)
		}
	}()

	wg.Wait()
	testWg.Wait()

	assert.Equal(t, expectedFilesAfterStop, counter)
	assert.Equal(t, expectedFilesAfterStop, cache.Len())
	assert.True(t, files.Empty())

	entries, err := os.ReadDir(dir)
	assert.NoError(t, err)
	assert.Len(t, entries, numFiles)
}
