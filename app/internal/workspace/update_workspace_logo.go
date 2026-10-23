package workspace

import (
	"context"
	"fmt"

	"selectDb/internal/api"
	"selectDb/internal/db/db_types"
	"selectDb/internal/db/generated"
	"selectDb/internal/edition"
)

// A pointer so a response missing the field reads as nil, not as an empty logo.
type logoResponse struct {
	Logo *string `json:"logo"`
}

// UpdateLogo uploads a logo, then mirrors what the server stored — the
// re-encoded image, not the caller's bytes — into the local database. Teammates
// get it on their next pull, since the endpoint bumps updated_at.
//
// The local write is untracked (@no-track in the query): the sync path does not
// carry the logo column, so a tracked one would queue a commit that never applies.
func (w *Workspace) UpdateLogo(workspaceID, logo string) error {
	ctx := context.Background()

	// No server to re-encode it. The frontend already draws it into a 128x128
	// PNG, and the column's CHECK refuses anything that is not one.
	if edition.Local {
		return w.storeLogoLocally(ctx, workspaceID, logo)
	}

	var res logoResponse
	if err := api.Fetch(ctx, "PUT", "workspaces/"+workspaceID+"/logo",
		map[string]string{"logo": logo},
		api.WorkspaceHeader(workspaceID), &res); err != nil {
		return fmt.Errorf("update workspace logo on server: %w", err)
	}
	if res.Logo == nil {
		return fmt.Errorf("update workspace logo on server: empty response")
	}

	return w.storeLogoLocally(ctx, workspaceID, *res.Logo)
}

func (w *Workspace) storeLogoLocally(ctx context.Context, workspaceID, logo string) error {
	if err := w.Queries.UpdateWorkspaceLogo(ctx, generated.UpdateWorkspaceLogoParams{
		ID:   workspaceID,
		Logo: db_types.NewJSONNullString(logo),
	}); err != nil {
		return fmt.Errorf("store workspace logo: %w", err)
	}

	// So the picker and settings panel show the new logo now, not at the next pull.
	if w.ReloadHooks != nil && w.ReloadHooks.BuildWorkspaceGraph != nil {
		if err := w.ReloadHooks.BuildWorkspaceGraph(); err != nil {
			return fmt.Errorf("rebuild workspace graph: %w", err)
		}
		if w.ReloadHooks.EmitWorkspaceGraphUpdated != nil {
			w.ReloadHooks.EmitWorkspaceGraphUpdated()
		}
	}
	return nil
}
