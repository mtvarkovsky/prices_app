package files

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestFileQueueInMem() *fileQueueInMem {
	files := NewFileQueueInMem(1)
	fq := files.(*fileQueueInMem)
	return fq
}

func TestFileQueueInMem_Put(t *testing.T) {
	files := newTestFileQueueInMem()
	err := files.Put(File{Path: "test"})
	assert.NoError(t, err)
}

func TestFileQueueInMem_Put_Error(t *testing.T) {
	files := newTestFileQueueInMem()
	close(files.data)
	err := files.Put(File{Path: "test"})
	assert.Error(t, err)
}

func TestFileQueueInMem_Data(t *testing.T) {
	files := newTestFileQueueInMem()
	_, err := files.Data()
	assert.NoError(t, err)
}

func TestFileQueueInMem_Get(t *testing.T) {
	files := newTestFileQueueInMem()
	testFile := File{Path: "test"}
	files.data <- testFile
	file, err := files.Get()
	assert.NoError(t, err)
	assert.Equal(t, testFile, file)
}

func TestFileQueueInMem_Get_Error(t *testing.T) {
	files := newTestFileQueueInMem()
	close(files.data)
	_, err := files.Get()
	assert.Error(t, err)
}

func TestFileQueueInMem_Close(t *testing.T) {
	files := newTestFileQueueInMem()
	err := files.Close()
	assert.NoError(t, err)
	_, ok := <-files.data
	assert.False(t, ok)
}

func TestFileQueueInMem_Close_Error(t *testing.T) {
	files := newTestFileQueueInMem()
	close(files.data)
	err := files.Close()
	assert.Error(t, err)
}

func TestFileQueueInMem_Empty_True(t *testing.T) {
	files := newTestFileQueueInMem()
	empty := files.Empty()
	assert.True(t, empty)
}

func TestFileQueueInMem_Empty_False(t *testing.T) {
	files := newTestFileQueueInMem()
	files.data <- File{Path: "test"}
	empty := files.Empty()
	assert.False(t, empty)
}

func newTestFileCacheInMem() *fileCacheInMem {
	cache := NewFileCacheInMem()
	fc := cache.(*fileCacheInMem)
	return fc
}

func TestFileCacheInMem_Put(t *testing.T) {
	cache := newTestFileCacheInMem()
	testFile := File{Path: "test"}
	err := cache.Put(testFile)
	assert.NoError(t, err)
	assert.Equal(t, testFile, cache.data[testFile.Path])
}

func TestFileCacheInMem_Get_Found(t *testing.T) {
	cache := newTestFileCacheInMem()
	testFile := File{Path: "test"}
	cache.data[testFile.Path] = testFile
	file, ok, err := cache.Get(testFile.Path)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, testFile, file)
}

func TestFileCacheInMem_Get_NotFound(t *testing.T) {
	cache := newTestFileCacheInMem()
	file, ok, err := cache.Get("test")
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Equal(t, File{}, file)
}

func TestFileCacheInMem_Len_NonEmpty(t *testing.T) {
	cache := newTestFileCacheInMem()
	testFile := File{Path: "test"}
	cache.data[testFile.Path] = testFile
	l := cache.Len()
	assert.Equal(t, 1, l)
}

func TestFileCacheInMem_Len_Empty(t *testing.T) {
	cache := newTestFileCacheInMem()
	l := cache.Len()
	assert.Equal(t, 0, l)
}
