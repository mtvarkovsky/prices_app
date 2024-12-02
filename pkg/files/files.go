package files

import (
	"fmt"
)

type (
	// File - object that represents file in the file system.
	File struct {
		// Path - path of the file
		Path string
		// Name - name of the file
		Name string
	}

	// FileQueue - interface a queue of File that are going to be processed
	FileQueue interface {
		// Put - adds a File to the queue
		Put(file File) error
		// Data - returns channel with File from queue
		Data() (<-chan File, error)
		// Get - returns a File from top of the queue.
		Get() (File, error)
		// Close - closes the queue
		Close() error
		// Empty - check if file queue is empty
		Empty() bool
	}

	// FileQueueInMem - in memory implementation of the FileQueue.
	// Only applicable for a single instance scanner per scanned directory, because multiple scanners can read same files multiple times.
	FileQueueInMem struct {
		data chan File
	}

	// FileCache - interface for File cache. It saves precessed File, so they would not be processed multiple times.
	FileCache interface {
		// Put - put File in cache
		Put(file File) error
		// Get - gets File from cache by its key.
		// Return File, bool value if file was found, and an error
		Get(key string) (File, bool, error)
		// Len - number of File in FileCache
		Len() int
	}

	// FileCacheInMem - in memory implementation of the FileCache.
	// Only applicable for a single instance scanner per scanned directory.
	// Not safe for concurrent usage.
	FileCacheInMem struct {
		// path -> File
		data map[string]File
	}
)

func (f File) String() string {
	return f.Path
}

func NewFileQueueInMem(size int) *FileQueueInMem {
	return &FileQueueInMem{
		data: make(chan File, size),
	}
}

func (q *FileQueueInMem) Put(file File) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("unable to put file=%s into queue: (%v)", file.Path, r)
		}
	}()
	q.data <- file
	return err
}

func (q *FileQueueInMem) Data() (<-chan File, error) {
	return q.data, nil
}

func (q *FileQueueInMem) Get() (File, error) {
	file, ok := <-q.data
	if !ok {
		return File{}, fmt.Errorf("uanble to get file from queue, queue is closed")
	}
	return file, nil
}

func (q *FileQueueInMem) Close() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("unable close files queue: (%v)", r)
		}
	}()
	close(q.data)
	return err
}

func (q *FileQueueInMem) Empty() bool {
	return len(q.data) == 0
}

func NewFileCacheInMem() *FileCacheInMem {
	return &FileCacheInMem{
		data: make(map[string]File),
	}
}

func (c *FileCacheInMem) Put(file File) error {
	c.data[file.Path] = file
	return nil
}

// Get - gets File from cache by path
func (c *FileCacheInMem) Get(path string) (File, bool, error) {
	file, ok := c.data[path]
	return file, ok, nil
}

func (c *FileCacheInMem) Len() int {
	return len(c.data)
}
