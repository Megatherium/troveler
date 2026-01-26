package lib

import (
	"os"
	"strings"
)

type OSInfo struct {
	ID        string
	Name      string
	Variant   string
	VariantID string
}

func DetectOS() (*OSInfo, error) {
	info := &OSInfo{}

	// Try /etc/os-release first (most modern distros)
	data, err := os.ReadFile("/etc/os-release")
	if err == nil {
		content := string(data)
		for _, line := range strings.Split(content, "\n") {
			if strings.HasPrefix(line, "ID=") {
				info.ID = strings.Trim(strings.TrimPrefix(line, "ID="), `"`)
			}
			if strings.HasPrefix(line, "NAME=") {
				info.Name = strings.Trim(strings.TrimPrefix(line, "NAME="), `"`)
			}
			if strings.HasPrefix(line, "VARIANT=") {
				info.Variant = strings.Trim(strings.TrimPrefix(line, "VARIANT="), `"`)
			}
			if strings.HasPrefix(line, "VARIANT_ID=") {
				info.VariantID = strings.Trim(strings.TrimPrefix(line, "VARIANT_ID="), `"`)
			}
		}
		if info.ID != "" {
			return normalizeOSInfo(info), nil
		}
	}

	// Try /etc/lsb-release (Ubuntu and derivatives)
	data, err = os.ReadFile("/etc/lsb-release")
	if err == nil {
		content := string(data)
		for _, line := range strings.Split(content, "\n") {
			if strings.HasPrefix(line, "DISTRIB_ID=") {
				info.ID = strings.Trim(strings.TrimPrefix(line, "DISTRIB_ID="), `"`)
			}
			if strings.HasPrefix(line, "DISTRIB_DESCRIPTION=") {
				info.Name = strings.Trim(strings.TrimPrefix(line, "DISTRIB_DESCRIPTION="), `"`)
			}
		}
		if info.ID != "" {
			return normalizeOSInfo(info), nil
		}
	}

	// Try /etc/redhat-release (RHEL/CentOS/Fedora)
	data, err = os.ReadFile("/etc/redhat-release")
	if err == nil {
		content := string(data)
		lower := strings.ToLower(content)
		if strings.Contains(lower, "fedora") {
			info.ID = "fedora"
			info.Name = "Fedora"
		} else if strings.Contains(lower, "centos") {
			info.ID = "centos"
			info.Name = "CentOS"
		} else if strings.Contains(lower, "rhel") || strings.Contains(lower, "red hat") {
			info.ID = "rhel"
			info.Name = "Red Hat Enterprise Linux"
		}
		return normalizeOSInfo(info), nil
	}

	// Try /etc/debian_version
	data, err = os.ReadFile("/etc/debian_version")
	if err == nil && len(data) > 0 {
		info.ID = "debian"
		info.Name = "Debian"
		return normalizeOSInfo(info), nil
	}

	// Try /etc/alpine-release
	_, err = os.ReadFile("/etc/alpine-release")
	if err == nil {
		info.ID = "alpine"
		info.Name = "Alpine Linux"
		return normalizeOSInfo(info), nil
	}

	// Try /etc/arch-release (Arch Linux)
	_, err = os.ReadFile("/etc/arch-release")
	if err == nil {
		info.ID = "arch"
		info.Name = "Arch Linux"
		return normalizeOSInfo(info), nil
	}

	// Try /etc/gentoo-release
	data, err = os.ReadFile("/etc/gentoo-release")
	if err == nil {
		info.ID = "gentoo"
		info.Name = "Gentoo"
		return normalizeOSInfo(info), nil
	}

	return info, nil
}

func normalizeOSInfo(info *OSInfo) *OSInfo {
	switch info.ID {
	case "ubuntu", "linuxmint", "pop":
		info.ID = "ubuntu"
	case "centos", "rhel", "rocky", "alma":
		info.ID = "rhel"
	case "fedora":
		info.ID = "fedora"
	case "debian":
		info.ID = "debian"
	case "arch", "manjaro", "endeavouros":
		info.ID = "arch"
	case "alpine":
		info.ID = "alpine"
	case "opensuse-tumbleweed", "opensuse-leap":
		info.ID = "opensuse"
	case "nixos":
		info.ID = "nixos"
	case "gentoo":
		info.ID = "gentoo"
	case "freebsd":
		info.ID = "freebsd"
	case "openbsd":
		info.ID = "openbsd"
	case "netbsd":
		info.ID = "netbsd"
	case "darwin":
		info.ID = "macos"
	case "windows":
		info.ID = "windows"
	}
	return info
}

func MatchPlatform(detectedID string, installPlatform string) bool {
	// Extract the distro part from platforms like "linux:fedora", "macos:brew"
	parts := strings.SplitN(installPlatform, ":", 2)
	platformOS := parts[0]
	platformMethod := ""
	if len(parts) > 1 {
		platformMethod = parts[1]
	}

	// For platform-only entries like "go", "rust", "eget"
	if platformOS == installPlatform {
		return platformOS == detectedID
	}

	// Check if the detected OS matches the platform OS (e.g., "macos:brew")
	if platformOS == detectedID {
		return true
	}

	// For "linux:*" entries, check if the method matches detected ID
	if platformOS == "linux" && platformMethod != "" {
		if platformMethod == detectedID {
			return true
		}
		// Check for distro aliases in method (e.g., "linux:ubuntu / debian")
		methods := strings.Split(platformMethod, "/")
		for _, m := range methods {
			m = strings.TrimSpace(m)
			if m == detectedID {
				return true
			}
		}
	}

	// For bsd entries, check if the method matches detected ID
	if platformOS == "bsd" && platformMethod != "" {
		if platformMethod == detectedID {
			return true
		}
	}

	// Check for rhel variants
	if detectedID == "rhel" {
		return platformOS == "rhel" || platformOS == "fedora" || platformOS == "centos"
	}

	// Check for arch variants
	if detectedID == "arch" {
		return platformOS == "arch" || platformOS == "manjaro"
	}

	// Check for ubuntu variants
	if detectedID == "ubuntu" {
		return platformOS == "ubuntu" || platformOS == "debian"
	}

	return false
}
