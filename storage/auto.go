package storage

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"log"
	"os"
	"path/filepath"
)

type AutoStorage struct {
	memoryStorage *MemoryStorage
	diskStorage   *DiskStorage
	maxMemorySize int64
}

func NewAutoStorage(maxMemorySize int64, diskPath string) *AutoStorage {
	return &AutoStorage{
		memoryStorage: NewMemoryStorage(),
		diskStorage:   NewDiskStorage(diskPath),
		maxMemorySize: maxMemorySize * 1024 * 1024, //To convert the memory size to MB to be compared with the file size
	}
}

func (as *AutoStorage) Save(key string, data []byte) error {
	if int64(len(data)) > as.maxMemorySize {
		return as.diskStorage.Save(key, data)
	}
	return as.memoryStorage.Save(key, data)
}

func (as *AutoStorage) Load(key string) ([]byte, error) {
	data, err := as.memoryStorage.Load(key)
	if err == nil {
		return data, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return as.diskStorage.Load(key)
	}
	return nil, err
}

func (as *AutoStorage) Delete(key string) error {
	err := as.memoryStorage.Delete(key)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return as.diskStorage.Delete(key)
}

func (as *AutoStorage) Size(key string) (int64, error) {
	size, err := as.memoryStorage.Size(key)
	if err == nil {
		return size, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return as.diskStorage.Size(key)
	}
	return 0, err
}

func (as *AutoStorage) Append(key string, data []byte) error {
	currentSize, err := as.Size(key)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return as.Save(key, data)
		}
		return err
	}

	if currentSize+int64(len(data)) > as.maxMemorySize {
		// Transition from memory storage to disk storage
		if as.IsInMemory(key) {

			existingData, err := as.memoryStorage.Load(key)
			if err != nil {
				return err
			}

			// Save existing data to disk
			err = as.diskStorage.Save(key, existingData)
			if err != nil {
				return err
			}
			// Delete from memory storage
			err = as.memoryStorage.Delete(key)
			if err != nil {
				return err
			}
		}
		// Append the new data to disk storage
		// Here, we ensure to append any remaining data after transitioning
		err = as.diskStorage.Append(key, data)
		if err != nil {
			return err
		}

		return nil
	}

	// If the combined size is within the maxMemorySize limit, append to memory
	return as.memoryStorage.Append(key, data)
}

func (as *AutoStorage) AppendFromReader(key string, r io.Reader) error {
	var buffer []byte
	currentSize, err := as.Size(key)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	// Read data from the io.Reader until the total read bytes exceed maxMemorySize
	for {
		chunk := make([]byte, 1024*1024) // 1MB buffer
		n, err := r.Read(chunk)
		if err != nil && err != io.EOF {
			return err
		}
		if n > 0 {
			buffer = append(buffer, chunk[:n]...)
			currentSize += int64(n)
			if currentSize > as.maxMemorySize {
				break
			}
		}
		if err == io.EOF {
			break
		}
	}

	// Save to memory if within the maxMemorySize limit
	if currentSize <= as.maxMemorySize {
		return as.memoryStorage.Save(key, buffer)
	}

	// Transition to disk storage
	if buffer != nil {
		// Save existing data to disk
		err = as.diskStorage.Save(key, buffer)
		if err != nil {
			return err
		}

		// Delete from memory storage
		err = as.memoryStorage.Delete(key)
		if err != nil {
			return err
		}
	}

	// Continue reading the remaining data and appending to disk storage
	_, err = io.Copy(&diskAppender{as.diskStorage, key}, r)
	if err != nil {
		return err
	}
	// Close the file after appending
	return as.diskStorage.CloseFile(key)
}

type diskAppender struct {
	diskStorage *DiskStorage
	key         string
}

func (da *diskAppender) Write(p []byte) (n int, err error) {
	err = da.diskStorage.appendToOpenFile(da.key, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// Add the IsInMemory method
func (as *AutoStorage) IsInMemory(key string) bool {
	_, err := as.memoryStorage.Load(key)
	if err == nil {
		return true
	}
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return false
}

func (as *AutoStorage) GetBasePath() string {
	return as.diskStorage.basePath
}

// Add the ComputeHash method to calculate and return the file hash
func (as *AutoStorage) ComputeHash(key string) (string, error) {
	isInMemory := as.IsInMemory(key)

	var hasher hash.Hash = sha256.New()

	if isInMemory {
		data, err := as.memoryStorage.Load(key)
		if err != nil {
			return "", err
		}
		_, err = hasher.Write(data)
		if err != nil {
			return "", err
		}
	} else {

		path := filepath.Join(as.diskStorage.basePath, key)
		file, err := os.Open(path)
		if err != nil {
			return "", err
		}
		defer file.Close()

		_, err = io.Copy(hasher, file)
		if err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// Add the Replace method to replace keyTo name to keyFrom then delete keyFrom
func (as *AutoStorage) ReplaceAndDelete(keyFrom, keyTo string) error {
	isInMemory := as.IsInMemory(keyFrom)
	if isInMemory {
		data, err := as.memoryStorage.Load(keyFrom)
		if err != nil {
			return err
		}
		err = as.Save(keyTo, data)
		if err != nil {
			return err
		}
		as.memoryStorage.Delete(keyFrom)
		return nil
	}
	// Check if newPath exists and delete it if it's a file
	keyToPath := filepath.Join(as.diskStorage.basePath, keyTo)
	_, err := os.Stat(keyToPath)
	if err == nil {
		if err := os.Remove(keyToPath); err != nil {
			log.Println(err)
		}
	}
	info, err := os.Stat(filepath.Join(as.diskStorage.basePath, keyFrom))
	if err != nil {
		return err
	}
	oldPath := filepath.Join(as.diskStorage.basePath, keyFrom)
	newPath := filepath.Join(as.diskStorage.basePath, keyTo)
	as.diskStorage.mutex.RLock()
	defer as.diskStorage.mutex.RUnlock()
	as.diskStorage.sizes[keyTo] = info.Size()
	return os.Rename(oldPath, newPath)

}

// ReadFileHeader reads the first 262 bytes of the file identified by the given key.
// If the file is smaller than 262 bytes, it reads the entire file.
func (as *AutoStorage) ReadFileHeader(key string) ([]byte, error) {
	const maxHeaderSize = 2048 // Maximum file header size to read

	var header []byte
	var err error

	if as.IsInMemory(key) {
		data, err := as.memoryStorage.Load(key)
		if err != nil {
			return nil, err
		}
		// Determine the number of bytes to read (min of 262 or data length)
		readLen := len(data)
		if readLen > maxHeaderSize {
			readLen = maxHeaderSize
		}
		header = data[:readLen]
	} else {
		path := filepath.Join(as.diskStorage.basePath, key)
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		// Allocate a buffer of 262 bytes
		buffer := make([]byte, maxHeaderSize)
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, err
		}
		header = buffer[:n]
	}

	return header, err
}

// Add the LoadAsReader method to AutoStorage
func (as *AutoStorage) LoadAsReader(key string) (io.ReadCloser, error) {
	data, err := as.memoryStorage.Load(key)
	if err == nil {
		// Use memory-backed reader if the file is in memory
		return io.NopCloser(bytes.NewReader(data)), nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return as.diskStorage.LoadAsReader(key)
	}
	return nil, err
}
