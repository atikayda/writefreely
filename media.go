/*
 * Copyright © 2025 Musing Studio LLC.
 *
 * This file is part of WriteFreely.
 *
 * WriteFreely is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License, included
 * in the LICENSE file in this source code package.
 */

package writefreely

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/writeas/web-core/log"
)

type remoteAvatarRequest struct {
	RemoteUserID int64  `json:"remote_user_id"`
	AvatarURL    string `json:"avatar_url"`
}

type remoteAvatarResponse struct {
	RemoteUserID int64  `json:"remote_user_id"`
	CachedURL    string `json:"cached_url"`
	StorageKey   string `json:"storage_key"`
}

func (app *App) cacheRemoteAvatar(remoteUserID int64, avatarURL string) (string, error) {
	if app.cfg.App.MediaServerURL == "" {
		return avatarURL, nil
	}

	if avatarURL == "" {
		return "", nil
	}

	endpoint := app.cfg.App.MediaServerURL + "/internal/remote-avatar"

	reqBody, err := json.Marshal(remoteAvatarRequest{
		RemoteUserID: remoteUserID,
		AvatarURL:    avatarURL,
	})
	if err != nil {
		return avatarURL, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return avatarURL, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if app.cfg.App.MediaServerAPIKey != "" {
		req.Header.Set("X-Internal-Key", app.cfg.App.MediaServerAPIKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Failed to cache remote avatar for user %d: %v", remoteUserID, err)
		return avatarURL, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("Media server returned status %d for remote avatar cache", resp.StatusCode)
		return avatarURL, fmt.Errorf("media server returned status %d", resp.StatusCode)
	}

	var result remoteAvatarResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return avatarURL, fmt.Errorf("failed to decode response: %w", err)
	}

	if err := app.db.UpdateRemoteUserAvatarCached(remoteUserID, result.CachedURL); err != nil {
		log.Error("Failed to update cached avatar URL in DB: %v", err)
	}

	return result.CachedURL, nil
}

func (app *App) updateRemoteUserAvatar(remoteUserID int64, displayName, avatarURL string) error {
	existingUser, err := app.db.GetRemoteUserWithAvatar(remoteUserID)
	if err != nil {
		return err
	}

	if existingUser.AvatarURL == avatarURL && existingUser.DisplayName == displayName {
		return nil
	}

	if err := app.db.UpdateRemoteUserProfile(remoteUserID, displayName, avatarURL); err != nil {
		return err
	}

	if avatarURL != "" && avatarURL != existingUser.AvatarURL {
		go func() {
			if _, err := app.cacheRemoteAvatar(remoteUserID, avatarURL); err != nil {
				log.Error("Failed to cache remote avatar: %v", err)
			}
		}()
	}

	return nil
}
