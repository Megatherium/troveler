package platform

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

	data, err = os.ReadFile("/etc/redhat-release")
	if err == nil {
		content := string(data)
		lower := strings.ToLower(content)
		if strings.Contains(lower, OSFedora) {
			info.ID = OSFedora
		} else if strings.Contains(lower, OSCentos) {
			info.ID = OSCentos
		} else if strings.Contains(lower, OSRHEL) || strings.Contains(lower, "red hat") {
			info.ID = OSRHEL
			info.Name = "Red Hat Enterprise Linux"
		}

		return normalizeOSInfo(info), nil
	}

	data, err = os.ReadFile("/etc/debian_version")
	if err == nil && len(data) > 0 {
		info.ID = OSDebian
		info.Name = "Debian"

		return normalizeOSInfo(info), nil
	}

	_, err = os.ReadFile("/etc/alpine-release")
	if err == nil {
		info.ID = OSAlpine
		info.Name = "Alpine Linux"

		return normalizeOSInfo(info), nil
	}

	_, err = os.ReadFile("/etc/arch-release")
	if err == nil {
		info.ID = OSArch
		info.Name = "Arch Linux"

		return normalizeOSInfo(info), nil
	}

	_, err = os.ReadFile("/etc/gentoo-release")
	if err == nil {
		info.ID = OSGentoo
		info.Name = "Gentoo"

		return normalizeOSInfo(info), nil
	}

	return info, nil
}

func normalizeOSInfo(info *OSInfo) *OSInfo {
	switch info.ID {
	case OSUbuntu, "linuxmint", "pop":
		info.ID = OSUbuntu
	case OSCentos, OSRHEL, "rocky", "alma":
		info.ID = OSRHEL
	case OSFedora:
		info.ID = OSFedora
	case OSAlpine:
		info.ID = OSAlpine
	case OSGentoo:
		info.ID = OSGentoo
	case OSDebian:
		info.ID = OSDebian
	case OSArch, OSManjaro, "endeavouros":
		info.ID = OSArch
	case "opensuse-tumbleweed", "opensuse-leap":
		info.ID = OSOpenSUSE
	case OSNixOS:
		info.ID = OSNixOS
	case OSFreeBSD:
		info.ID = OSFreeBSD
	case OSOpenBSD:
		info.ID = OSOpenBSD
	case OSNetBSD:
		info.ID = OSNetBSD
	case OSDarwin:
		info.ID = OSMacOS
	case OSWindows:
		info.ID = OSWindows
	}

	return info
}
