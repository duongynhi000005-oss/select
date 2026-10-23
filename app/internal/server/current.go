package server

import (
	"os"
	"path/filepath"
	"strings"

	"selectDb/internal/edition"
	"selectDb/internal/utils"
)

const currentServerFileName = ".current_server"

// LocalDomain names the local edition's one server. Nothing listens there: it
// keys the local database and the workspace configs, and nothing else.
const LocalDomain = "local"

// currentServerFilePath returns the path to the file that stores the current server domain.
func currentServerFilePath() (string, error) {
	appRoot, err := utils.GetAppDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(appRoot, currentServerFileName), nil
}

// ReadCurrentDomain returns the current server domain from the file, or empty if unset.
// The local edition ignores the file, which a pro install on the same machine
// shares.
func ReadCurrentDomain() (string, error) {
	if edition.Local {
		return LocalDomain, nil
	}
	path, err := currentServerFilePath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// WriteCurrentDomain writes the current server domain to the file.
func WriteCurrentDomain(domain string) error {
	path, err := currentServerFilePath()
	if err != nil {
		return err
	}
	domain = strings.TrimSpace(domain)
	return os.WriteFile(path, []byte(domain), 0o600)
}
