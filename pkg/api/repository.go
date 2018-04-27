package api

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	mongodb  = "mongodb"
	file     = "file"
	postgres = "postgres"
)

// Repository defines the behavior of a proxy specs repository
type Repository interface {
	io.Closer
	FindAll() ([]*Definition, error)
}

// Watcher defines how a provider should watch for changes on configurations
type Watcher interface {
	Watch(ctx context.Context, cfgChan chan<- ConfigurationChanged)
}

// Listener defines how a provider should listen for changes on configurations
type Listener interface {
	Listen(ctx context.Context, cfgChan <-chan ConfigurationMessage)
}

// BuildRepository creates a repository instance that will depend on your given DSN
func BuildRepository(dsn string, refreshTime time.Duration) (Repository, error) {
	dsnURL, err := url.Parse(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing the DSN")
	}

	switch dsnURL.Scheme {
	case mongodb:
		log.Debug("MongoDB configuration chosen")
		return NewMongoAppRepository(dsn, refreshTime)
	case file:
		log.Debug("File system based configuration chosen")
		apiPath := fmt.Sprintf("%s/apis", dsnURL.Path)

		log.WithField("api_path", apiPath).Debug("Trying to load configuration files")
		repo, err := NewFileSystemRepository(apiPath)
		if err != nil {
			return nil, errors.Wrap(err, "could not create a file system repository")
		}
		return repo, nil
	case postgres:
		log.Debug("Postgres configuration chosen")
		return NewPostgresRepository(dsn)
	default:
		return nil, errors.New("The selected scheme is not supported to load API definitions")
	}
}
