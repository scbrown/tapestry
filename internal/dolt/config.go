package dolt

import (
	"fmt"
	"strings"
)

// Config holds Dolt server connection parameters.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
}

// DefaultConfig returns a Config with common defaults for Dolt.
func DefaultConfig() Config {
	return Config{
		Host: "127.0.0.1",
		Port: 3306,
		User: "root",
	}
}

// Validate checks that required fields are set.
func (c Config) Validate() error {
	if strings.TrimSpace(c.Host) == "" {
		return fmt.Errorf("dolt: host is required")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("dolt: port must be between 1 and 65535, got %d", c.Port)
	}
	if strings.TrimSpace(c.User) == "" {
		return fmt.Errorf("dolt: user is required")
	}
	return nil
}

// DSN returns a go-sql-driver/mysql data source name.
// It connects without a default database so we can USE different databases.
func (c Config) DSN() string {
	auth := c.User
	if c.Password != "" {
		auth = c.User + ":" + c.Password
	}
	return fmt.Sprintf("%s@tcp(%s:%d)/?parseTime=true&multiStatements=true&readTimeout=10s&writeTimeout=10s&timeout=5s", auth, c.Host, c.Port)
}
