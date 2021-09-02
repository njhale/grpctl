package internal

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	yaml "gopkg.in/yaml.v2"
)

type FS interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, mode fs.FileMode) error
	RemoveFile(name string) error
	ReadDir(name string) ([]fs.DirEntry, error)
}

var _ FS = &subFS{}

type subFS struct {
	root string
}

func (s *subFS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(s.rooted(name))
}

func (s *subFS) WriteFile(name string, data []byte, mode fs.FileMode) error {
	return os.WriteFile(s.rooted(name), data, mode)
}

func (s *subFS) RemoveFile(name string) error {
	return os.Remove(s.rooted(name))
}

func (s *subFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(s.rooted(name))
}

func (s *subFS) rooted(name string) string {
	return filepath.Join(s.root, name)
}

func DefaultConfigFS() (*subFS, error) {
	configFS := &subFS{
		root: os.Getenv("XDG_CONFIG_HOME"),
	}
	if configFS.root == "" {
		home, err := homedir.Dir()
		if err != nil {
			return nil, err
		}
		configFS.root = filepath.Join(home, ".grpctl")
	}

	return configFS, os.MkdirAll(configFS.root, 0o774)
}

func NewServerFileStore(filesystem FS) *ServerFileStore {
	return &ServerFileStore{
		fs: filesystem,
	}
}

type ServerFileStore struct {
	fs FS

	// TODO(njhale): DI Marshal/Unmarshal
}

func (s *ServerFileStore) Set(server Server) error {
	marshalled, err := yaml.Marshal(&server)
	if err != nil {
		return fmt.Errorf("Failed to set config for server %s: %w", server.Name, err)
	}

	return s.fs.WriteFile(server.Name, marshalled, 0o774)
}

func (s *ServerFileStore) Get(name string) (Server, error) {
	var unmarshalled Server
	marshalled, err := s.fs.ReadFile(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// Usually means that no servers have been configured yet
			return unmarshalled, nil
		}

		return unmarshalled, fmt.Errorf("Failed to get config for server %s: %w", name, err)
	}

	return unmarshalled, yaml.Unmarshal(marshalled, &unmarshalled)
}

func (s *ServerFileStore) Remove(name string) error {
	if err := s.fs.RemoveFile(name); err != nil {
		return fmt.Errorf("Failed to remove config for server: %s: %w", name, err)
	}

	return nil
}

type errorList []error

func (e errorList) Wrapped() error {
	if len(e) < 1 {
		return nil
	}

	return fmt.Errorf(strings.Repeat("%w ", len(e)-1)+"%w", e)
}

func (s *ServerFileStore) List() ([]Server, error) {
	entries, err := s.fs.ReadDir(".")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// Usually means that no servers have been configured yet
			return nil, nil
		}

		return nil, fmt.Errorf("Failed to list configured servers: %w", err)
	}

	var (
		servers []Server
		errs    errorList
	)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		server, err := s.Get(entry.Name())
		if err != nil {
			errs = append(errs, err)
			continue
		}

		servers = append(servers, server)
	}

	return servers, errs.Wrapped()
}
