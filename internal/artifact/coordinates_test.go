package artifact

import "testing"

func TestValidateIdentifier(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid identifier",
			input:   "abc",
			wantErr: false,
		},
		{
			name:    "valid with separators",
			input:   "my-org.api_v2",
			wantErr: false,
		},
		{
			name:    "valid with digits",
			input:   "artifact123",
			wantErr: false,
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "uppercase",
			input:   "MyOrg",
			wantErr: true,
		},
		{
			name:    "whitespace",
			input:   "my org",
			wantErr: true,
		},
		{
			name:    "slash",
			input:   "my/org",
			wantErr: true,
		},
		{
			name:    "leading separator",
			input:   "-myorg",
			wantErr: true,
		},
		{
			name:    "trailing separator",
			input:   "myorg.",
			wantErr: true,
		},
		{
			name:    "unicode",
			input:   "jájjj",
			wantErr: true,
		},
		{
			name:    "path traversal",
			input:   "..",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIdentifier(tt.input)

			if tt.wantErr && err == nil {
				t.Errorf("validateIdentifier(%q) expected an error, got nil", tt.input)
			}

			if !tt.wantErr && err != nil {
				t.Errorf("validateIdentifier(%q) unexpected error: %v", tt.input, err)
			}
		})
	}
}
func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid stable version",
			input:   "1.0.0",
			wantErr: false,
		},
		{
			name:    "valid zero minor version",
			input:   "0.1.0",
			wantErr: false,
		},
		{
			name:    "valid prerelease version",
			input:   "1.2.3-alpha.1",
			wantErr: false,
		},
		{
			name:    "valid build metadata",
			input:   "1.2.3+build.42",
			wantErr: false,
		},
		{
			name:    "leading v",
			input:   "v1.0.0",
			wantErr: true,
		},
		{
			name:    "missing patch version",
			input:   "1.0",
			wantErr: true,
		},
		{
			name:    "extra version component",
			input:   "1.0.0.0",
			wantErr: true,
		},
		{
			name:    "latest tag",
			input:   "latest",
			wantErr: true,
		},
		{
			name:    "empty version",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVersion(tt.input)

			if tt.wantErr && err == nil {
				t.Errorf("validateVersion(%q) expected an error, got nil", tt.input)
			}

			if !tt.wantErr && err != nil {
				t.Errorf("validateVersion(%q) unexpected error: %v", tt.input, err)
			}
		})
	}
}

func TestNewCoordinates(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		artifact  string
		version   string
		wantErr   bool
	}{
		{
			name:      "valid coordinates",
			namespace: "acme",
			artifact:  "api",
			version:   "1.0.0",
			wantErr:   false,
		},
		{
			name:      "invalid namespace",
			namespace: "Acme",
			artifact:  "api",
			version:   "1.0.0",
			wantErr:   true,
		},
		{
			name:      "invalid artifact name",
			namespace: "acme",
			artifact:  "api/",
			version:   "1.0.0",
			wantErr:   true,
		},
		{
			name:      "invalid version",
			namespace: "acme",
			artifact:  "api",
			version:   "latest",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coords, err := NewCoordinates(tt.namespace, tt.artifact, tt.version)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if coords.namespace != tt.namespace {
					t.Errorf("namespace = %q, want %q", coords.namespace, tt.namespace)
				}

				if coords.name != tt.artifact {
					t.Errorf("name = %q, want %q", coords.name, tt.artifact)
				}

				if coords.version != tt.version {
					t.Errorf("version = %q, want %q", coords.version, tt.version)
				}
			}
		})
	}
}
