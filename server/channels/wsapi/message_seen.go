// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package wsapi

import (
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/request"
)

func (api *API) InitMessageSeen() {
	api.Router.Handle("message_seen", api.APIWebSocketHandler(api.messageSeen))
}

func (api *API) messageSeen(req *model.WebSocketRequest) (map[string]any, *model.AppError) {
	var ok bool
	var postID string
	if postID, ok = req.Data["post_id"].(string); !ok || !model.IsValidId(postID) {
		return nil, NewInvalidWebSocketParamError(req.Action, "post_id")
	}

	rctx := request.EmptyContext(api.App.Log())

	if !api.App.SessionHasPermissionToChannelByPost(req.Session, postID, model.PermissionReadChannel) {
		return nil, NewInvalidWebSocketParamError(req.Action, "post_id")
	}

	receipt, appErr := api.App.MarkMessageAsSeen(rctx, postID, req.Session.UserId)
	if appErr != nil {
		return nil, appErr
	}

	if receipt == nil {
		return nil, nil
	}

	return map[string]any{
		"post_id": receipt.PostID,
		"user_id": receipt.UserID,
		"seen_at": receipt.SeenAt,
	}, nil
}

