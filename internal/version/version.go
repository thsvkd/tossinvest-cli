package version

// Repo is the canonical GitHub "owner/repo" slug — the single source of truth
// for every runtime reference to this project's repository (release links,
// update-check API, schema URL). Change it here once and they all follow.
//
// NOTE: Go *import paths* (github.com/JungHoonGhae/tossinvest-cli/...) cannot
// use this — import paths are compile-time string literals fixed by the
// language, so a repo rename still requires a module-path find-replace across
// all files. This constant only centralizes runtime URLs.
const Repo = "JungHoonGhae/tossinvest-cli"

// ReleasesLatestURL is the human-facing "latest release" page.
const ReleasesLatestURL = "https://github.com/" + Repo + "/releases/latest"

// RawMainURL builds a raw.githubusercontent.com URL for a file on the main
// branch (e.g. RawMainURL("schemas/config.schema.json")).
func RawMainURL(path string) string {
	return "https://raw.githubusercontent.com/" + Repo + "/main/" + path
}

type Info struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date,omitempty"`
}

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = ""
)

func Current() Info {
	return Info{
		Version: Version,
		Commit:  Commit,
		Date:    Date,
	}
}
