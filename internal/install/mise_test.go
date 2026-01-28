package install

import (
	"testing"
)

func TestTransformToMise(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "go install github.com",
			input:    "go install github.com/user/repo",
			expected: "mise use --global go:github.com/user/repo",
		},
		{
			name:     "go install https://github.com",
			input:    "go install https://github.com/user/repo",
			expected: "mise use --global go:github.com/user/repo",
		},
		{
			name:     "cargo install",
			input:    "cargo install crate-name",
			expected: "mise use --global cargo:crate-name",
		},
		{
			name:     "npm install -g",
			input:    "npm install -g package-name",
			expected: "mise use --global npm:package-name",
		},
		{
			name:     "yarn global add",
			input:    "yarn global add package-name",
			expected: "mise use --global npm:package-name",
		},
		{
			name:     "pnpm add -g",
			input:    "pnpm add -g package-name",
			expected: "mise use --global npm:package-name",
		},
		{
			name:     "npm install (no -g)",
			input:    "npm install package-name",
			expected: "mise use --global npm:package-name",
		},
		{
			name:     "pip install",
			input:    "pip install package-name",
			expected: "mise use --global pipx:package-name",
		},
		{
			name:     "pip3 install",
			input:    "pip3 install package-name",
			expected: "mise use --global pipx:package-name",
		},
		{
			name:     "pipx install",
			input:    "pipx install package-name",
			expected: "mise use --global pipx:package-name",
		},
		{
			name:     "brew install (unsupported)",
			input:    "brew install package-name",
			expected: "brew install package-name",
		},
		{
			name:     "apt install (unsupported)",
			input:    "apt install package-name",
			expected: "apt install package-name",
		},
		{
			name:     "extra whitespace",
			input:    "  go install github.com/user/repo  ",
			expected: "mise use --global go:github.com/user/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TransformToMise(tt.input)
			if result != tt.expected {
				t.Errorf("TransformToMise(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
