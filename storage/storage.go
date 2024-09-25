package storage

import "io"

type StorageClient interface {
	Save(key string, data []byte) error
	Load(key string) ([]byte, error)
	Delete(key string) error
	Size(key string) (int64, error)
	Append(key string, data []byte) error
	AppendFromReader(key string, r io.Reader) error
	IsInMemory(key string) bool
	GetBasePath() string
	ComputeHash(key string) (string, error)
	ReplaceAndDelete(keyFrom, keyTo string) error
	ReadFileHeader(key string) ([]byte, error)
	LoadAsReader(key string) (io.ReadCloser, error)
}
