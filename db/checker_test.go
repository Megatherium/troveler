package db

import (
	"testing"
)

// parseToolNameTests covers every resolver with real-world command strings
// sourced from the actual troveler database.
var parseToolNameTests = []struct {
	command string
	want    string
}{
	// --- Go install ---
	{"go install github.com/dhth/act3@latest", "act3"},
	{"go install github.com/arimxyer/aic@latest", "aic"},
	{"go install github.com/nakabonne/ali@latest", "ali"},
	{"go install filippo.io/age/cmd/...@latest", "age"},
	{"go install -ldflags='-s -w' github.com/tjblackheart/andcli/v2/cmd/andcli@latest", "andcli"},
	{"go install github.com/IAL32/az-tui/cmd/az-tui@latest", "az-tui"},
	{"go install github.com/superstarryeyes/bit/cmd/bit@latest", "bit"},
	{"go install github.com/LeperGnome/bt/cmd/bt@v1.0.0", "bt"},
	{"go install github.com/edoardottt/cariddi/cmd/cariddi@latest", "cariddi"},
	{"go install github.com/pressly/goose/v3/cmd/goose@latest", "goose"},
	{"go install https://github.com/rubysolo/brows@latest", "brows"},
	{"go install github.com/ekkinox/yai@latest", "yai"},

	// --- Cargo ---
	{"cargo install adguardian", "adguardian"},
	{"cargo install ad-editor", "ad-editor"},
	{"cargo install --locked bacon", "bacon"},
	{"cargo binstall ast-grep", "ast-grep"},
	{"cargo install --git https://github.com/boxdot/gurk-rs gurk", "gurk"},
	{"cargo install --locked --git https://github.com/pamburus/hl.git", "hl"},
	{"cargo install --git https://github.com/fioncat/otree", "otree"},
	{"cargo install --git=https://github.com/nate-sys/tuime", "tuime"},
	{"cargo install --locked cargo-seek", "cargo-seek"},
	{"cargo install --locked cargo-geiger", "cargo-geiger"},

	// --- npm ---
	{"npm install --global @ast-grep/cli", "cli"},
	{"npm install -g @augmentcode/auggie", "auggie"},
	{"npm install -g branchlet", "branchlet"},
	{"npm i -g carbon-now-cli", "carbon-now-cli"},
	{"npm install -g @anthropic-ai/claude-code", "claude-code"},
	{"npm i -g @openai/codex", "codex"},
	{"npm install daff -g", "daff"},
	{"npm install -g @fresh-editor/fresh-editor", "fresh-editor"},
	{"npm install -g @google/gemini-cli", "gemini-cli"},
	{"npm install -g git-split-diffs", "git-split-diffs"},
	{"npm install -g httpyac", "httpyac"},
	{"npm install -g mapscii", "mapscii"},
	{"npm install -g @jdxcode/mise", "mise"},
	{"npm install -g neoss", "neoss"},
	{"npm i -g @thekarel/rum", "rum"},
	{"npm install -g tldr", "tldr"},
	{"npm install -g vtop", "vtop"},
	{"npm install -g openclaw@latest", "openclaw"},
	{"npm install -g ralph-tui", "ralph-tui"},
	{"npm install -g @googleworkspace/cli", "cli"},
	{"npm install -g workos", "workos"},
	{"npm install -g deadbranch", "deadbranch"},

	// --- yarn / pnpm ---
	{"yarn global add carbon-now-cli", "carbon-now-cli"},
	{"pnpm add -g openclaw@latest", "openclaw"},

	// --- pip ---
	{"pip install ast-grep-cli", "ast-grep-cli"},
	{"pip install bkp", "bkp"},
	{"pip install blueutil-tui", "blueutil-tui"},
	{"pip install braindrop", "braindrop"},
	{"pip3 install buku", "buku"},
	{"pip install calcure", "calcure"},
	{"python -m pip install aider-install", "aider-install"},
	{"python3 -m pip install aria2tui", "aria2tui"},

	// --- pipx ---
	{"pipx install aider-chat", "aider-chat"},
	{"pipx install austin-dist", "austin-dist"},
	{"pipx install blueutil-tui", "blueutil-tui"},
	{"pipx install braindrop", "braindrop"},
	{"pipx install browsr", "browsr"},
	{"pipx install calcure", "calcure"},
	{"pipx install celerator", "celerator"},

	// --- uv tool ---
	{"uv tool install --force --python python3.12 --with pip aider-chat@latest", "aider-chat"},
	{"uv tool install --python 3.13 bagels", "bagels"},
	{"uv tool install blueutil-tui", "blueutil-tui"},
	{"uv tool install celerator", "celerator"},
	{"uv tool install cloctui", "cloctui"},
	{"uv tool install --python 3.13 dhv", "dhv"},
	{"uv tool install dirsearch", "dirsearch"},
	{"uv tool install dotbins", "dotbins"},

	// --- uvx ---
	{"uvx some-tool", "some-tool"},

	// --- rye ---
	{"rye install some-tool", "some-tool"},

	// --- Homebrew ---
	{"brew install aerc", "aerc"},
	{"brew install ad", "ad"},
	{"brew install age", "age"},
	{"brew install bat", "bat"},
	{"linuxbrew install curl", "curl"},

	// --- apt ---
	{"apt install age", "age"},
	{"apt-get install aerc", "aerc"},

	// --- pacman ---
	{"pacman -S aerc", "aerc"},
	{"pacman -S ad", "ad"},

	// --- dnf / yum ---
	{"dnf install age", "age"},
	{"yum install something", "something"},

	// --- zypper ---
	{"zypper install age", "age"},
	{"zypper install aerc", "aerc"},

	// --- apk ---
	{"apk add aerc", "aerc"},
	{"apk add age", "age"},
	{"apk add curl", "curl"},

	// --- emerge ---
	{"emerge install aerc", "aerc"},
	{"emerge app-crypt/age", "age"},
	{"emerge net-misc/aria2", "aria2"},
	{"emerge app-misc/asciinema", "asciinema"},

	// --- FreeBSD ---
	{"pkg_add aerc", "aerc"},
	{"pkgin install aerc", "aerc"},
	{"pkg install aerc", "aerc"},

	// --- MacPorts ---
	{"port install aerc", "aerc"},
	{"sudo port install aerc", "aerc"},
	{"sudo port install ali", "ali"},
	{"sudo port install bat", "bat"},

	// --- Nix ---
	{"nix-env -iA nixos.aerc", "aerc"},
	{"nix-env -iA ad", "ad"},
	{"nix-env -iA adguardian", "adguardian"},
	{"nix-env -iA nixpkgs.aichat", "aichat"},
	{"nix-env -i age", "age"},
	{"nix-shell -p asn", "asn"},

	// --- Scoop ---
	{"scoop install extras/adguardian", "adguardian"},
	{"scoop install extras/age", "age"},
	{"scoop install aichat", "aichat"},

	// --- Chocolatey ---
	{"choco install age.portable", "age.portable"},
	{"choco install Aria2 Client", "Aria2"},

	// --- winget ---
	{"winget install Some.App", "Some.App"},

	// --- Snap ---
	{"sudo snap install bandwhich", "bandwhich"},
	{"sudo snap install btop", "btop"},

	// --- eget ---
	{"eget dhth/act3", "act3"},
	{"eget Lissy93/AdGuardian-Term", "AdGuardian-Term"},
	{"eget FiloSottile/age", "age"},
	{"eget nakabonne/ali", "ali"},
	{"eget makew0rld/amfora", "amfora"},
	{"eget sharkdp/bat", "bat"},
	{"eget atuinsh/atuin", "atuin"},

	// --- Arch AUR helpers ---
	{"yay -S somepackage", "somepackage"},
	{"paru -Syu adguardian", "adguardian"},
	{"paru -S something", "something"},

	// --- mise ---
	{"mise use --global npm:package", "package"},
	{"mise install node:22", "22"},

	// --- Gem ---
	{"gem install somegem", "somegem"},

	// --- Cabal ---
	{"cabal install something", "something"},

	// --- npx ---
	{"npx some-tool", "some-tool"},

	// --- pkgman ---
	{"pkgman install somepkg", "somepkg"},

	// --- sudo prefixed ---
	{"sudo apt install age", "age"},
	{"sudo apt-get install aerc", "aerc"},
	{"sudo dnf install something", "something"},
	{"sudo ports install broot", "broot"},

	// --- curl/wget scripts ---
	{"curl -fsSL https://claude.ai/install.sh | bash", "claude"},
	{"curl -fsSL https://cli.coderabbit.ai/install.sh | sh", "cli"},

	// --- Edge cases: leading whitespace (TrimSpace handles it) ---
	{" pip3 install castero", "castero"},
	{" cargo install --locked rucola-notes", "rucola-notes"},

	// --- Edge cases ---
	{"", ""},    // empty
	{"   ", ""}, // whitespace only
}

func TestParseToolName(t *testing.T) {
	for _, tt := range parseToolNameTests {
		t.Run(tt.command, func(t *testing.T) {
			got := parseToolName(tt.command)
			if got != tt.want {
				t.Errorf("parseToolName(%q) = %q, want %q", tt.command, got, tt.want)
			}
		})
	}
}

func TestParseToolName_ScopedNpmPackages(t *testing.T) {
	tests := []struct {
		command string
		want    string
	}{
		{"npm install --global @ast-grep/cli", "cli"},
		{"npm install -g @augmentcode/auggie", "auggie"},
		{"npm install -g @anthropic-ai/claude-code", "claude-code"},
		{"npm i -g @openai/codex", "codex"},
		{"npm install -g @fresh-editor/fresh-editor", "fresh-editor"},
		{"npm install -g @google/gemini-cli", "gemini-cli"},
		{"npm install -g @jdxcode/mise", "mise"},
		{"npm i -g @thekarel/rum", "rum"},
		{"npm install -g @googleworkspace/cli", "cli"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			got := parseToolName(tt.command)
			if got != tt.want {
				t.Errorf("parseToolName(%q) = %q, want %q", tt.command, got, tt.want)
			}
		})
	}
}

func TestParseToolName_GoInstall(t *testing.T) {
	tests := []struct {
		command string
		want    string
	}{
		{"go install github.com/dhth/act3@latest", "act3"},
		{"go install github.com/arimxyer/aic@latest", "aic"},
		{"go install filippo.io/age/cmd/...@latest", "age"},
		{"go install -ldflags='-s -w' github.com/tjblackheart/andcli/v2/cmd/andcli@latest", "andcli"},
		{"go install github.com/IAL32/az-tui/cmd/az-tui@latest", "az-tui"},
		{"go install github.com/superstarryeyes/bit/cmd/bit@latest", "bit"},
		{"go install github.com/LeperGnome/bt/cmd/bt@v1.0.0", "bt"},
		{"go install github.com/pressly/goose/v3/cmd/goose@latest", "goose"},
		{"go install https://github.com/rubysolo/brows@latest", "brows"},
		{"go install github.com/ekkinox/yai@latest", "yai"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			got := parseToolName(tt.command)
			if got != tt.want {
				t.Errorf("parseToolName(%q) = %q, want %q", tt.command, got, tt.want)
			}
		})
	}
}

func TestParseToolName_Eget(t *testing.T) {
	tests := []struct {
		command string
		want    string
	}{
		{"eget dhth/act3", "act3"},
		{"eget Lissy93/AdGuardian-Term", "AdGuardian-Term"},
		{"eget FiloSottile/age", "age"},
		{"eget sharkdp/bat", "bat"},
		{"eget atuinsh/atuin", "atuin"},
		{"eget orhun/binsider", "binsider"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			got := parseToolName(tt.command)
			if got != tt.want {
				t.Errorf("parseToolName(%q) = %q, want %q", tt.command, got, tt.want)
			}
		})
	}
}

func TestResolveExecutableName_PrefersOverride(t *testing.T) {
	inst := InstallInstruction{
		Command:        "npm install -g @scope/some-pkg",
		ExecutableName: "actual-binary",
	}
	got := resolveExecutableName(inst)
	if got != "actual-binary" {
		t.Errorf("resolveExecutableName with ExecutableName set = %q, want %q", got, "actual-binary")
	}
}

func TestResolveExecutableName_FallsBackToParsing(t *testing.T) {
	inst := InstallInstruction{
		Command: "brew install bat",
	}
	got := resolveExecutableName(inst)
	if got != "bat" {
		t.Errorf("resolveExecutableName without ExecutableName = %q, want %q", got, "bat")
	}
}

func TestCleanName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"strips version", "package@1.2.3", "package"},
		{"strips @latest", "package@latest", "package"},
		{"extracts last path component", "github.com/user/repo", "repo"},
		{"no slash no at", "simple", "simple"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanName(tt.input)
			if got != tt.want {
				t.Errorf("cleanName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsInstalled_NilTool(t *testing.T) {
	if IsInstalled(nil, []InstallInstruction{{Command: "brew install bat"}}) {
		t.Error("IsInstalled(nil, ...) should return false")
	}
}

func TestIsInstalled_EmptyInstalls(t *testing.T) {
	tool := &Tool{ID: "test"}
	if IsInstalled(tool, nil) {
		t.Error("IsInstalled(tool, nil) should return false")
	}
	if IsInstalled(tool, []InstallInstruction{}) {
		t.Error("IsInstalled(tool, []) should return false")
	}
}

func TestIsInstalled_UsesExecutableName(t *testing.T) {
	tool := &Tool{ID: "test"}
	// "go" is almost certainly available on this system
	installs := []InstallInstruction{
		{Command: "npm install -g @scope/wrong-name", ExecutableName: "go"},
	}
	if !IsInstalled(tool, installs) {
		t.Error("IsInstalled should return true when ExecutableName resolves to an available command")
	}
}

func TestIsInstalled_FallsBackToParsing(t *testing.T) {
	tool := &Tool{ID: "test"}
	// "go" is available on this system, and brew install go resolves to "go"
	installs := []InstallInstruction{
		{Command: "brew install go"},
	}
	if !IsInstalled(tool, installs) {
		t.Error("IsInstalled should return true when parsed command resolves to an available command")
	}
}

func TestBuildLookPathCache_DeduplicatesNames(t *testing.T) {
	installsByTool := map[string][]InstallInstruction{
		"tool-1": {
			{Command: "brew install go"},
			{Command: "cargo install ripgrep"},
		},
		"tool-2": {
			{Command: "brew install go"},        // duplicate name "go"
			{Command: "npm install -g eslint"},  // unique name "eslint"
		},
	}

	cache := BuildLookPathCache(installsByTool)

	// Should have exactly 3 unique names: go, ripgrep, eslint
	if len(cache) != 3 {
		t.Errorf("expected 3 unique names, got %d", len(cache))
	}

	// "go" is available on this system
	if !cache["go"] {
		t.Error("expected 'go' to be in cache as available")
	}
}

func TestBuildLookPathCache_EmptyInput(t *testing.T) {
	cache := BuildLookPathCache(nil)
	if len(cache) != 0 {
		t.Errorf("expected empty cache for nil input, got %d entries", len(cache))
	}

	cache = BuildLookPathCache(map[string][]InstallInstruction{})
	if len(cache) != 0 {
		t.Errorf("expected empty cache for empty map, got %d entries", len(cache))
	}
}

func TestIsInstalledCached_CacheHit(t *testing.T) {
	tool := &Tool{ID: "test"}
	installs := []InstallInstruction{
		{Command: "brew install mytool"},
	}
	cache := map[string]bool{"mytool": true}

	if !IsInstalledCached(tool, installs, cache) {
		t.Error("expected true when cache has executable as available")
	}
}

func TestIsInstalledCached_CacheMiss(t *testing.T) {
	tool := &Tool{ID: "test"}
	installs := []InstallInstruction{
		{Command: "brew install absent-tool"},
	}
	cache := map[string]bool{"absent-tool": false}

	if IsInstalledCached(tool, installs, cache) {
		t.Error("expected false when cache has executable as unavailable")
	}
}

func TestIsInstalledCached_NilTool(t *testing.T) {
	cache := map[string]bool{"go": true}
	if IsInstalledCached(nil, []InstallInstruction{{Command: "brew install go"}}, cache) {
		t.Error("expected false for nil tool")
	}
}

func TestIsInstalledCached_EmptyInstalls(t *testing.T) {
	tool := &Tool{ID: "test"}
	cache := map[string]bool{"go": true}
	if IsInstalledCached(tool, nil, cache) {
		t.Error("expected false for nil installs")
	}
	if IsInstalledCached(tool, []InstallInstruction{}, cache) {
		t.Error("expected false for empty installs")
	}
}

func TestIsInstalledCached_UsesExecutableName(t *testing.T) {
	tool := &Tool{ID: "test"}
	installs := []InstallInstruction{
		{Command: "npm install -g @scope/wrong-name", ExecutableName: "my-real-bin"},
	}
	cache := map[string]bool{"my-real-bin": true}

	if !IsInstalledCached(tool, installs, cache) {
		t.Error("expected true when ExecutableName matches cache")
	}
}
