package db

import (
	"context"
	"fmt"
)

// SaveToolSnapshot atomically refreshes a tool and replaces its install
// instructions. Existing tool IDs and tags are preserved. The supplied tool
// is updated with its stored identity only after the transaction commits.
func (s *SQLiteDB) SaveToolSnapshot(ctx context.Context, tool *Tool, installs []InstallInstruction) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tool refresh: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	storedTool := *tool
	if err := upsertTool(ctx, tx, &storedTool); err != nil {
		return fmt.Errorf("refresh tool %q: %w", tool.Slug, err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM install_instructions WHERE tool_id = ?", storedTool.ID); err != nil {
		return fmt.Errorf("replace installs for %q: %w", tool.Slug, err)
	}
	for _, inst := range installs {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO install_instructions (id, tool_id, platform, command, executable_name)
			VALUES (?, ?, ?, ?, ?)`,
			inst.ID, storedTool.ID, inst.Platform, inst.Command, inst.ExecutableName,
		)
		if err != nil {
			return fmt.Errorf("save install for %q on %q: %w", tool.Slug, inst.Platform, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tool refresh %q: %w", tool.Slug, err)
	}
	*tool = storedTool
	return nil
}
