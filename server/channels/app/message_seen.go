// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
	"github.com/mattermost/mattermost/server/public/shared/request"
	"github.com/mattermost/mattermost/server/v8/channels/store"
)

func (a *App) MarkMessageAsSeen(rctx request.CTX, postID, userID string) (*model.ReadReceipt, *model.AppError) {
	post, err := a.GetSinglePost(rctx, postID, false)
	if err != nil {
		return nil, err
	}

	// Don't mark own messages as seen
	if post.UserId == userID {
		return nil, nil
	}

	channel, err := a.GetChannel(rctx, post.ChannelId)
	if err != nil {
		return nil, err
	}

	if channel.DeleteAt > 0 {
		return nil, model.NewAppError("MarkMessageAsSeen", "api.message_seen.archived_channel.app_error", nil, "", http.StatusForbidden)
	}

	receipt := &model.ReadReceipt{
		PostID:    postID,
		UserID:    userID,
		ChannelID: post.ChannelId,
	}

	savedReceipt, nErr := a.Srv().Store().ReadReceipt().Save(rctx, receipt)
	if nErr != nil {
		var appErr *model.AppError
		if errors.As(nErr, &appErr) {
			return nil, appErr
		}
		return nil, model.NewAppError("MarkMessageAsSeen", "app.message_seen.save.app_error", nil, "", http.StatusInternalServerError).Wrap(nErr)
	}

	a.sendMessageSeenEvent(rctx, savedReceipt, post)

	return savedReceipt, nil
}

func (a *App) GetSeenUsersForPost(rctx request.CTX, postID string, limit, offset int) ([]*model.ReadReceipt, *model.AppError) {
	receipts, err := a.Srv().Store().ReadReceipt().GetByPostWithUsers(rctx, postID, limit, offset)
	if err != nil {
		return nil, model.NewAppError("GetSeenUsersForPost", "app.message_seen.get.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}
	return receipts, nil
}

func (a *App) GetSeenCountForPost(rctx request.CTX, postID string) (int64, *model.AppError) {
	count, err := a.Srv().Store().ReadReceipt().GetReadCountForPost(rctx, postID)
	if err != nil {
		return 0, model.NewAppError("GetSeenCountForPost", "app.message_seen.count.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}
	return count, nil
}

func (a *App) GetSeenReceiptsForPosts(rctx request.CTX, postIDs []string) (map[string][]*model.ReadReceipt, *model.AppError) {
	receipts, err := a.Srv().Store().ReadReceipt().GetForPosts(rctx, postIDs)
	if err != nil {
		return nil, model.NewAppError("GetSeenReceiptsForPosts", "app.message_seen.get_for_posts.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	receiptsMap := make(map[string][]*model.ReadReceipt)
	for _, r := range receipts {
		receiptsMap[r.PostID] = append(receiptsMap[r.PostID], r)
	}

	return receiptsMap, nil
}

func (a *App) HasUserSeenPost(rctx request.CTX, postID, userID string) (bool, *model.AppError) {
	_, err := a.Srv().Store().ReadReceipt().Get(rctx, postID, userID)
	if err != nil {
		var nfErr *store.ErrNotFound
		if errors.As(err, &nfErr) {
			return false, nil
		}
		return false, model.NewAppError("HasUserSeenPost", "app.message_seen.check.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}
	return true, nil
}

func (a *App) sendMessageSeenEvent(rctx request.CTX, receipt *model.ReadReceipt, post *model.Post) {
	message := model.NewWebSocketEvent(model.WebsocketEventMessageSeen, "", post.ChannelId, "", nil, "")

	receiptJSON, err := json.Marshal(receipt)
	if err != nil {
		rctx.Logger().Warn("Failed to encode read receipt to JSON", mlog.Err(err))
		return
	}
	message.Add("read_receipt", string(receiptJSON))
	message.Add("post_id", post.Id)
	a.Publish(message)
}

