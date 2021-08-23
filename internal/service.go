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

type Service struct {
	Name    string
	Address string
}

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

func NewServiceFileStore(filesystem FS) *ServiceFileStore {
	return &ServiceFileStore{
		fs: filesystem,
	}
}

type ServiceFileStore struct {
	fs FS

	// TODO(njhale): DI Marshal/Unmarshal
}

func (s *ServiceFileStore) Set(service Service) error {
	marshalled, err := yaml.Marshal(&service)
	if err != nil {
		return fmt.Errorf("Failed to set config for service %s: %w", service.Name, err)
	}

	return s.fs.WriteFile(service.Name, marshalled, 0o774)
}

func (s *ServiceFileStore) Get(name string) (Service, error) {
	var unmarshalled Service
	marshalled, err := s.fs.ReadFile(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// Usually means that no services have been configured yet
			return unmarshalled, nil
		}

		return unmarshalled, fmt.Errorf("Failed to get config for service %s: %w", name, err)
	}

	return unmarshalled, yaml.Unmarshal(marshalled, &unmarshalled)
}

func (s *ServiceFileStore) Remove(name string) error {
	if err := s.fs.RemoveFile(name); err != nil {
		return fmt.Errorf("Failed to remove config for service: %s: %w", name, err)
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

func (s *ServiceFileStore) List() ([]Service, error) {
	entries, err := s.fs.ReadDir(".")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// Usually means that no services have been configured yet
			return nil, nil
		}

		return nil, fmt.Errorf("Failed to list configured services: %w", err)
	}

	var (
		services []Service
		errs     errorList
	)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		service, err := s.Get(entry.Name())
		if err != nil {
			errs = append(errs, err)
			continue
		}

		services = append(services, service)
	}

	return services, errs.Wrapped()
}
