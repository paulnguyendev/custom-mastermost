// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

type ReadReceipt struct {
	PostID    string `json:"post_id"`
	UserID    string `json:"user_id"`
	ChannelID string `json:"channel_id"`
	SeenAt    int64  `json:"seen_at"`
	ExpireAt  int64  `json:"expire_at"`
}

// IsValid validates the ReadReceipt fields
func (r *ReadReceipt) IsValid() *AppError {
	if !IsValidId(r.PostID) {
		return NewAppError("ReadReceipt.IsValid", "model.read_receipt.is_valid.post_id.app_error", nil, "", 400)
	}
	if !IsValidId(r.UserID) {
		return NewAppError("ReadReceipt.IsValid", "model.read_receipt.is_valid.user_id.app_error", nil, "", 400)
	}
	return nil
}

// PreSave sets default values before saving
func (r *ReadReceipt) PreSave() {
	if r.SeenAt == 0 {
		r.SeenAt = GetMillis()
	}
}
