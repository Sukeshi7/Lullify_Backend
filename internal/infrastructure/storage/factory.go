package storage

import (
	"fmt"
	"strings"
)

type Options struct {
	Provider  string
	LocalPath string
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

func New(opts Options) (Storage, error) {
	switch strings.ToLower(strings.TrimSpace(opts.Provider)) {
	case "local", "filesystem", "disk":
		return NewLocalStorage(opts.LocalPath)
	case "minio", "s3":
		return NewMinIOClient(opts.Endpoint, opts.AccessKey, opts.SecretKey, opts.Bucket, opts.UseSSL)
	default:
		return nil, fmt.Errorf("unknown storage provider %q (expected \"local\" or \"minio\"/\"s3\")", opts.Provider)
	}
}