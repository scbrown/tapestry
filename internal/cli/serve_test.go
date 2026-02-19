package cli

import "testing"

func TestServeCmd_DefaultFlags(t *testing.T) {
	cmd := newServeCmd()

	tests := []struct {
		flag string
		want string
	}{
		{"host", "localhost"},
		{"port", "8070"},
		{"dolt-host", "127.0.0.1"},
		{"dolt-port", "3306"},
		{"dolt-user", "root"},
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			f := cmd.Flags().Lookup(tt.flag)
			if f == nil {
				t.Fatalf("flag %q not found", tt.flag)
			}
			if f.DefValue != tt.want {
				t.Errorf("flag %q default = %q, want %q", tt.flag, f.DefValue, tt.want)
			}
		})
	}
}
