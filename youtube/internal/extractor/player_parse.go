// Adapted from github.com/kkdai/youtube, commit 3876742b95fd1cd96f836f396c2d54be8912a297.
// Copyright (c) 2015 Evan Lin. MIT license: see LICENSE and THIRD_PARTY_NOTICES.md.
package extractor

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"
)

type playerConfig []byte

var basejsPattern = regexp.MustCompile(`(/s/player/[A-Za-z0-9_-]+/[A-Za-z0-9._/-]*/base\.js)`)

func (c *Client) getPlayerConfig(ctx context.Context, videoID string) (playerConfig, error) {
	if c.playerCache.expiredAt.After(time.Now()) && len(c.playerCache.config) > 0 {
		return c.playerCache.config, nil
	}
	embedURL := fmt.Sprintf("https://youtube.com/embed/%s?hl=en", videoID)
	embedBody, err := c.httpGetBodyBytes(ctx, embedURL)
	if err != nil {
		return nil, err
	}

	// example: /s/player/f676c671/player_ias.vflset/en_US/base.js
	playerPath := string(basejsPattern.Find(embedBody))
	if playerPath == "" {
		return nil, errors.New("unable to find basejs URL in playerConfig")
	}

	config := c.playerCache.Get(playerPath)
	if config != nil {
		return config, nil
	}

	config, err = c.httpGetBodyBytes(ctx, "https://youtube.com"+playerPath)
	if err != nil {
		return nil, err
	}

	c.playerCache.Set(playerPath, config)
	return config, nil
}
