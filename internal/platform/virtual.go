package platform

import (
	"strings"

	"troveler/db"
)

type Backend string

const (
	BackendGo     Backend = "go"
	BackendCargo  Backend = "cargo"
	BackendNPM    Backend = "npm"
	BackendPipx   Backend = "pipx"
	BackendGithub Backend = "github"
)

type VirtualInstall struct {
	Platform string
	Command  string
}

func ResolveVirtual(platform string) string {
	if strings.HasPrefix(platform, "mise:") {
		return strings.TrimPrefix(platform, "mise:")
	}

	return platform
}

func GenerateVirtualInstalls(installs []db.InstallInstruction) []VirtualInstall {
	virtuals := make(map[Backend]VirtualInstall)

	for _, inst := range installs {
		if strings.HasPrefix(inst.Platform, "mise:") {
			continue
		}

		transformed := TransformToMise(inst.Command)

		if transformed == inst.Command {
			continue
		}

		backend := extractBackend(transformed)
		if backend == "" {
			continue
		}

		virtuals[backend] = VirtualInstall{
			Platform: "mise:" + string(backend),
			Command:  transformed,
		}
	}

	result := make([]VirtualInstall, 0, len(virtuals))
	order := []Backend{BackendGo, BackendCargo, BackendNPM, BackendPipx, BackendGithub}
	for _, backend := range order {
		if v, exists := virtuals[backend]; exists {
			result = append(result, v)
		}
	}

	return result
}

func FilterDBInstalls(installs []db.InstallInstruction, platform string, toolLanguage string) ([]db.InstallInstruction, bool) {
	var matched []db.InstallInstruction

	if platform == "LANG" {
		for _, inst := range installs {
			if MatchLanguage(toolLanguage, inst.Platform) {
				matched = append(matched, inst)
			}
		}
	} else {
		normalizedPlatform := Normalize(platform)
		for _, inst := range installs {
			if MatchPlatform(normalizedPlatform, inst.Platform) {
				matched = append(matched, inst)
			}
		}
	}

	if len(matched) == 0 {
		return installs, true
	}

	return matched, false
}

func SelectDefaultDBInstalls(installs []db.InstallInstruction, usedFallback bool, detectedOS string) *db.InstallInstruction {
	if len(installs) == 0 {
		return nil
	}

	if usedFallback {
		if detectedOS != "" {
			detectedLower := strings.ToLower(detectedOS)
			for _, inst := range installs {
				platformLower := strings.ToLower(inst.Platform)
				if platformLower == detectedLower {
					return &inst
				}
				if strings.HasSuffix(platformLower, ":"+detectedLower) {
					return &inst
				}
				if strings.HasSuffix(detectedLower, ":"+platformLower) {
					return &inst
				}
			}
		}

		var preferredPlatforms []string

		if IsLinuxFamily(detectedOS) {
			preferredPlatforms = []string{
				"linux:brew", "brew",
				"apt", "apt-get", "linux:ubuntu", "linux:debian",
				"pacman", "linux:arch", "linux:manjaro",
				"dnf", "linux:fedora", "yum", "linux:rhel", "linux:centos",
			}
		} else if IsMacFamily(detectedOS) {
			preferredPlatforms = []string{"macos:brew", "brew", "macos:macports", "macos"}
		} else if IsBSDFamily(detectedOS) {
			preferredPlatforms = []string{"bsd:freebsd", "bsd:openbsd", "bsd:netbsd", "bsd", "brew"}
		} else {
			preferredPlatforms = []string{"brew", "winget", "chocolatey", "scoop"}
		}

		for _, preferred := range preferredPlatforms {
			for _, inst := range installs {
				if strings.Contains(strings.ToLower(inst.Platform), preferred) {
					return &inst
				}
			}
		}

		return nil
	}

	return &installs[0]
}

func extractBackend(miseCommand string) Backend {
	if !strings.HasPrefix(miseCommand, "mise use --global ") {
		return ""
	}

	rest := strings.TrimPrefix(miseCommand, "mise use --global ")

	parts := strings.SplitN(rest, ":", 2)
	if len(parts) < 2 {
		return ""
	}

	backend := parts[0]

	switch Backend(backend) {
	case BackendGo, BackendCargo, BackendNPM, BackendPipx, BackendGithub:
		return Backend(backend)
	default:
		return ""
	}
}
