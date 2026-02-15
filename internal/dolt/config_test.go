package dolt

import "testing"

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Host != "127.0.0.1" {
		t.Errorf("Host = %q, want %q", cfg.Host, "127.0.0.1")
	}
	if cfg.Port != 3306 {
		t.Errorf("Port = %d, want %d", cfg.Port, 3306)
	}
	if cfg.User != "root" {
		t.Errorf("User = %q, want %q", cfg.User, "root")
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "valid default",
			cfg:     DefaultConfig(),
			wantErr: false,
		},
		{
			name:    "valid with password",
			cfg:     Config{Host: "dolt.svc", Port: 3306, User: "tapestry", Password: "secret"},
			wantErr: false,
		},
		{
			name:    "empty host",
			cfg:     Config{Host: "", Port: 3306, User: "root"},
			wantErr: true,
		},
		{
			name:    "whitespace host",
			cfg:     Config{Host: "  ", Port: 3306, User: "root"},
			wantErr: true,
		},
		{
			name:    "port zero",
			cfg:     Config{Host: "localhost", Port: 0, User: "root"},
			wantErr: true,
		},
		{
			name:    "port too high",
			cfg:     Config{Host: "localhost", Port: 70000, User: "root"},
			wantErr: true,
		},
		{
			name:    "empty user",
			cfg:     Config{Host: "localhost", Port: 3306, User: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_DSN(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "no password",
			cfg:  Config{Host: "127.0.0.1", Port: 3306, User: "root"},
			want: "root@tcp(127.0.0.1:3306)/?parseTime=true&multiStatements=true",
		},
		{
			name: "with password",
			cfg:  Config{Host: "dolt.svc", Port: 3307, User: "tapestry", Password: "secret"},
			want: "tapestry:secret@tcp(dolt.svc:3307)/?parseTime=true&multiStatements=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.DSN()
			if got != tt.want {
				t.Errorf("DSN() = %q, want %q", got, tt.want)
			}
		})
	}
}
