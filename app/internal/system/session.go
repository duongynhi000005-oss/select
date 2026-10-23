package system

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os/user"

	"selectDb/internal/api"
	"selectDb/internal/db/db_types"
	"selectDb/internal/db/generated"
	"selectDb/internal/edition"
	"selectDb/internal/graph"
	"selectDb/internal/utils"

	"selectDb/internal/desktop"
)

func (s *System) Logout() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = api.ClearAccessToken()
	_ = api.ClearRefreshToken()

	// The next person to sign in here is not necessarily the last one.
	s.closeOpenFolder()

	desktop.Emit("logout")
	return nil
}

func (s *System) closeOpenFolder() {
	graph.ClearOpenWorkspace()
	if s.Graph != nil {
		s.Graph.InvalidateWorkspaceGraph()
	}
	if s.fileWatcherCancel != nil {
		s.fileWatcherCancel()
		s.fileWatcherCancel = nil
	}
}

// CheckForLogout ends the session only on credentials the keyring reports as
// gone. A keyring that cannot answer is not a sign-out.
func (s *System) CheckForLogout() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if edition.Local {
		return
	}
	if api.ReadCredentialStatus() == api.CredentialsMissing {
		s.closeOpenFolder()
		desktop.Emit("logout")
	}
}

func (s *System) CheckForLogin() {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := s.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	// No event in the local edition: the frontend signs in when this returns.
	if edition.Local {
		if err := s.ensureLocalUser(ctx); err != nil {
			log.Printf("local user: %v", err)
		}
		return
	}

	if api.ReadCredentialStatus() != api.CredentialsPresent {
		return
	}

	_, err := s.Queries.GetCurrentUser(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_ = api.ClearAccessToken()
			_ = api.ClearRefreshToken()
		}
		return
	}
	desktop.Emit("login")
}

// ensureLocalUser stands in for the sign-in the local edition does not have:
// everything keyed on the current user then works as it does after one.
func (s *System) ensureLocalUser(ctx context.Context) error {
	_, err := s.Queries.GetCurrentUser(ctx)
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	name := "Local"
	if u, err := user.Current(); err == nil {
		switch {
		case u.Name != "":
			name = u.Name
		case u.Username != "":
			name = u.Username
		}
	}

	created, err := s.Queries.UpsertUser(ctx, generated.UpsertUserParams{
		ID:   utils.GenerateUUID(),
		Name: db_types.NewJSONNullString(name),
	})
	if err != nil {
		return err
	}
	return s.Queries.SetCurrentUser(ctx, created.ID)
}
