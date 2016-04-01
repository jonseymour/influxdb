package monitor

import (
	"time"

	"github.com/influxdata/influxdb/toml"
)

const (
	// DefaultStoreEnabled is whether the system writes gathered information in
	// an InfluxDB system for historical analysis.
	DefaultStoreEnabled = true

	// DefaultStoreDatabase is the name of the database where gathered information is written
	DefaultStoreDatabase = "_internal"

	// DefaultStoreInterval is the period between storing gathered information.
	DefaultStoreInterval = 10 * time.Second

	// DefaultIdleTime is the maximum time an idle Statistics object can reach before it is ignored.
	DefaultIdleTime = time.Hour // time.Hour
)

// Config represents the configuration for the monitor service.
type Config struct {
	StoreEnabled  bool          `toml:"store-enabled"`
	StoreDatabase string        `toml:"store-database"`
	StoreInterval toml.Duration `toml:"store-interval"`
	IdleTime      toml.Duration `toml:"idle-time"`
}

// NewConfig returns an instance of Config with defaults.
func NewConfig() Config {
	return Config{
		StoreEnabled:  true,
		StoreDatabase: DefaultStoreDatabase,
		StoreInterval: toml.Duration(DefaultStoreInterval),
		IdleTime:      toml.Duration(DefaultIdleTime),
	}
}
