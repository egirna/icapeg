package storage

import (
	"io"
	"os"
	"path/filepath"
	"sync"
)

type DiskStorage struct {
	basePath string
	sizes    map[string]int64
	mutex    sync.RWMutex
	openFile map[string]*os.File
}

func NewDiskStorage(basePath string) *DiskStorage {
	return &DiskStorage{
		basePath: basePath,
		sizes:    make(map[string]int64),
		openFile: make(map[string]*os.File),
	}
}

func (ds *DiskStorage) Save(key string, data []byte) error {
	path := filepath.Join(ds.basePath, key)
	err := os.WriteFile(path, data, 0644)
	if err == nil {
		ds.mutex.Lock()
		ds.sizes[key] = int64(len(data))
		ds.mutex.Unlock()
	}
	return err
}

func (ds *DiskStorage) Load(key string) ([]byte, error) {
	path := filepath.Join(ds.basePath, key)
	return os.ReadFile(path)
}

func (ds *DiskStorage) Delete(key string) error {
	path := filepath.Join(ds.basePath, key)
	err := os.Remove(path)
	if err == nil {
		ds.mutex.Lock()
		delete(ds.sizes, key)
		ds.mutex.Unlock()
	}
	return err
}

func (ds *DiskStorage) Size(key string) (int64, error) {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()
	size, exists := ds.sizes[key]
	if !exists {
		return 0, os.ErrNotExist
	}
	return size, nil
}

func (ds *DiskStorage) Append(key string, data []byte) error {
	path := filepath.Join(ds.basePath, key)

	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// If the file does not exist, create it and save the data
		return ds.Save(key, data)
	}

	// If the file exists, append the data
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err == nil {
		ds.mutex.Lock()
		ds.sizes[key] += int64(len(data))
		ds.mutex.Unlock()
	}
	return err
}

func (ds *DiskStorage) AppendFromReader(key string, r io.Reader) error {
	path := filepath.Join(ds.basePath, key)

	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// If the file does not exist, create it
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	// Open the file in append mode
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Append data from the reader
	written, err := io.Copy(file, r)
	if err == nil {
		ds.mutex.Lock()
		ds.sizes[key] += written
		ds.mutex.Unlock()
	}
	return err
}
func (ds *DiskStorage) appendToOpenFile(key string, data []byte) error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	path := filepath.Join(ds.basePath, key)

	// Check if the file is already open
	file, exists := ds.openFile[key]
	if !exists {
		// Open the file in append mode
		var err error
		file, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		ds.openFile[key] = file
	}

	_, err := file.Write(data)
	if err == nil {
		ds.sizes[key] += int64(len(data))
	}
	return err
}

func (ds *DiskStorage) CloseFile(key string) error {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	file, exists := ds.openFile[key]
	if !exists {
		return nil
	}

	err := file.Close()
	if err != nil {
		return err
	}

	delete(ds.openFile, key)
	return nil
}
func (ds *DiskStorage) LoadAsReader(key string) (io.ReadCloser, error) {
	path := filepath.Join(ds.basePath, key)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return file, nil // Return io.ReadCloser to stream the file
}
