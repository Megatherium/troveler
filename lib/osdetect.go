package lib

import (
	"os"
	"strings"
)

const (
	osIDFedora = "fedora"
	osIDCentos = "centos"
	osIDAlpine = "alpine"
	osIDGentoo = "gentoo"
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
		if strings.Contains(lower, osIDFedora) {
			info.ID = osIDFedora
		} else if strings.Contains(lower, osIDCentos) {
			info.ID = osIDCentos
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
		info.ID = osIDAlpine
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
	_, err = os.ReadFile("/etc/gentoo-release")
	if err == nil {
		info.ID = osIDGentoo
		info.Name = "Gentoo"
		return normalizeOSInfo(info), nil
	}

	return info, nil
}

func normalizeOSInfo(info *OSInfo) *OSInfo {
	switch info.ID {
	case "ubuntu", "linuxmint", "pop":
		info.ID = "ubuntu"
	case osIDCentos, "rhel", "rocky", "alma":
		info.ID = "rhel"
	case osIDFedora:
		info.ID = osIDFedora
	case osIDAlpine:
		info.ID = osIDAlpine
	case osIDGentoo:
		info.ID = osIDGentoo
	case "debian":
		info.ID = "debian"
	case "arch", "manjaro", "endeavouros":
		info.ID = "arch"
	case "opensuse-tumbleweed", "opensuse-leap":
		info.ID = "opensuse"
	case "nixos":
		info.ID = "nixos"
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
