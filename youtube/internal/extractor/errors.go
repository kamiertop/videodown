// Adapted from github.com/kkdai/youtube, commit 3876742b95fd1cd96f836f396c2d54be8912a297.
// Copyright (c) 2015 Evan Lin. MIT license: see LICENSE and THIRD_PARTY_NOTICES.md.
package extractor

import (
	"fmt"
)

const (
	ErrCipherNotFound             = constError("cipher not found")
	ErrSignatureTimestampNotFound = constError("signature timestamp not found")
	ErrInvalidCharactersInVideoID = constError("invalid characters in video id")
	ErrVideoIDMinLength           = constError("the video id must be at least 10 characters long")
	ErrReadOnClosedResBody        = constError("http: read on closed response body")
	ErrNotPlayableInEmbed         = constError("embedding of this video has been disabled")
	ErrLoginRequired              = constError("YouTube 要求登录或验证当前网络")
	ErrVideoPrivate               = constError("user restricted access to this video")
)

type constError string

func (e constError) Error() string {
	return string(e)
}

type ErrPlayabiltyStatus struct {
	Status string
	Reason string
}

func (err ErrPlayabiltyStatus) Error() string {
	return fmt.Sprintf("cannot playback and download, status: %s, reason: %s", err.Status, err.Reason)
}

// ErrUnexpectedStatusCode is returned on unexpected HTTP status codes
type ErrUnexpectedStatusCode int

func (err ErrUnexpectedStatusCode) Error() string {
	return fmt.Sprintf("unexpected status code: %d", err)
}
