/*
 * Copyright © 2025 Musing Studio LLC.
 *
 * This file is part of WriteFreely.
 *
 * WriteFreely is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License, included
 * in the LICENSE file in this source code package.
 */

package migrations

func supportRemoteInteractions(db *datastore) error {
	t, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = t.Exec(`CREATE TABLE remote_interactions (
	id               ` + db.typeIntPrimaryKey() + `,
	type             ` + db.typeVarChar(10) + ` NOT NULL,
	post_id          ` + db.typeChar(16) + ` NOT NULL,
	remote_user_id   ` + db.typeInt() + ` NOT NULL,
	activity_id      ` + db.typeVarChar(512) + ` NOT NULL,
	parent_id        ` + db.typeInt() + ` NULL,
	level            ` + db.typeTinyInt() + ` NOT NULL DEFAULT 1,
	content_markdown ` + db.typeText() + ` NULL,
	content_html     ` + db.typeText() + ` NULL,
	cw_text          ` + db.typeVarChar(500) + ` NULL,
	url              ` + db.typeVarChar(512) + ` NOT NULL,
	created          ` + db.typeDateTime() + ` NOT NULL,
	received         ` + db.typeDateTime() + ` NOT NULL
)` + db.engine())
	if err != nil {
		t.Rollback()
		return err
	}

	if db.driverName == driverMySQL {
		_, err = t.Exec(`CREATE UNIQUE INDEX idx_interactions_activity ON remote_interactions (activity_id)`)
		if err != nil {
			t.Rollback()
			return err
		}

		_, err = t.Exec(`CREATE INDEX idx_interactions_post ON remote_interactions (post_id, type, level, created)`)
		if err != nil {
			t.Rollback()
			return err
		}

		_, err = t.Exec(`CREATE INDEX idx_interactions_parent ON remote_interactions (parent_id)`)
		if err != nil {
			t.Rollback()
			return err
		}
	}

	_, err = t.Exec(`ALTER TABLE remoteusers ADD COLUMN display_name ` + db.typeVarChar(255) + ` NULL`)
	if err != nil {
		t.Rollback()
		return err
	}

	_, err = t.Exec(`ALTER TABLE remoteusers ADD COLUMN avatar_url ` + db.typeVarChar(512) + ` NULL`)
	if err != nil {
		t.Rollback()
		return err
	}

	_, err = t.Exec(`ALTER TABLE remoteusers ADD COLUMN avatar_cached ` + db.typeVarChar(255) + ` NULL`)
	if err != nil {
		t.Rollback()
		return err
	}

	_, err = t.Exec(`ALTER TABLE remoteusers ADD COLUMN updated ` + db.typeDateTime() + ` NULL`)
	if err != nil {
		t.Rollback()
		return err
	}

	err = t.Commit()
	if err != nil {
		t.Rollback()
		return err
	}

	return nil
}
