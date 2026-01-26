package lib

import (
	"testing"
)

func TestMatchPlatform(t *testing.T) {
	tests := []struct {
		name            string
		detectedID      string
		installPlatform string
		want            bool
	}{
		{
			name:            "fedora matches linux:fedora",
			detectedID:      "fedora",
			installPlatform: "linux:fedora",
			want:            true,
		},
		{
			name:            "fedora does not match linux:arch",
			detectedID:      "fedora",
			installPlatform: "linux:arch",
			want:            false,
		},
		{
			name:            "pure go matches",
			detectedID:      "go",
			installPlatform: "go",
			want:            true,
		},
		{
			name:            "pure rust doesn't match go",
			detectedID:      "rust",
			installPlatform: "go",
			want:            false,
		},
		{
			name:            "method with distro aliases",
			detectedID:      "ubuntu",
			installPlatform: "linux:ubuntu / debian",
			want:            true,
		},
		{
			name:            "macos matches",
			detectedID:      "macos",
			installPlatform: "macos:brew",
			want:            true,
		},
		{
			name:            "freebsd matches",
			detectedID:      "freebsd",
			installPlatform: "bsd:freebsd",
			want:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchPlatform(tt.detectedID, tt.installPlatform)
			if got != tt.want {
				t.Errorf("MatchPlatform(%q, %q) = %v, want %v", tt.detectedID, tt.installPlatform, got, tt.want)
			}
		})
	}
}

func TestNormalizeOSInfo(t *testing.T) {
	tests := []struct {
		name   string
		input  *OSInfo
		wantID string
	}{
		{
			name:   "ubuntu variants",
			input:  &OSInfo{ID: "linuxmint"},
			wantID: "ubuntu",
		},
		{
			name:   "rhel variants",
			input:  &OSInfo{ID: "centos"},
			wantID: "rhel",
		},
		{
			name:   "arch variants",
			input:  &OSInfo{ID: "manjaro"},
			wantID: "arch",
		},
		{
			name:   "fedora stays fedora",
			input:  &OSInfo{ID: "fedora"},
			wantID: "fedora",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeOSInfo(tt.input)
			if got.ID != tt.wantID {
				t.Errorf("normalizeOSInfo(%+v).ID = %v, want %v", tt.input, got.ID, tt.wantID)
			}
		})
	}
}
