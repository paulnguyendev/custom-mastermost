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
	// Send WebSocket event only to the post author
	message := model.NewWebSocketEvent(model.WebsocketEventMessageSeen, "", "", post.UserId, nil, "")

	receiptJSON, err := json.Marshal(receipt)
	if err != nil {
		rctx.Logger().Warn("Failed to encode read receipt to JSON", mlog.Err(err))
		return
	}
	message.Add("read_receipt", string(receiptJSON))
	message.Add("post_id", post.Id)
	a.Publish(message)
}

// MarkUnreadMessagesAsSeen marks all unread messages in a channel as seen by the user
// This is called when a user views a channel (from web or mobile app)
func (a *App) MarkUnreadMessagesAsSeen(rctx request.CTX, channelID, userID string, lastViewedAt int64) *model.AppError {
	// Get posts created after lastViewedAt (unread posts)
	options := model.GetPostsSinceOptions{
		ChannelId:       channelID,
		Time:            lastViewedAt,
		SkipFetchThreads: true,
		CollapsedThreads: false,
		UserId:          userID,
	}

	postList, err := a.Srv().Store().Post().GetPostsSince(rctx, options, false, a.Config().GetSanitizeOptions())
	if err != nil {
		rctx.Logger().Warn("Failed to get posts since lastViewedAt for marking as seen",
			mlog.String("channel_id", channelID),
			mlog.String("user_id", userID),
			mlog.Err(err))
		return nil // Don't fail the view operation
	}

	if postList == nil || len(postList.Posts) == 0 {
		return nil
	}

	// Filter posts: only mark posts from other users as seen
	var receiptsToSave []*model.ReadReceipt
	var postsToNotify []*model.Post
	for _, post := range postList.Posts {
		// Skip own posts and deleted posts
		if post.UserId == userID || post.DeleteAt > 0 {
			continue
		}
		// Skip system messages
		if post.IsSystemMessage() {
			continue
		}

		receiptsToSave = append(receiptsToSave, &model.ReadReceipt{
			PostID:    post.Id,
			UserID:    userID,
			ChannelID: channelID,
		})
		postsToNotify = append(postsToNotify, post)
	}

	if len(receiptsToSave) == 0 {
		return nil
	}

	// Batch save read receipts
	savedReceipts, saveErr := a.Srv().Store().ReadReceipt().SaveMultiple(rctx, receiptsToSave)
	if saveErr != nil {
		rctx.Logger().Warn("Failed to batch save read receipts",
			mlog.String("channel_id", channelID),
			mlog.String("user_id", userID),
			mlog.Err(saveErr))
		return nil // Don't fail the view operation
	}

	// Send WebSocket events to post authors
	for i, receipt := range savedReceipts {
		if i < len(postsToNotify) {
			a.sendMessageSeenEvent(rctx, receipt, postsToNotify[i])
		}
	}

	rctx.Logger().Debug("Marked messages as seen",
		mlog.String("channel_id", channelID),
		mlog.String("user_id", userID),
		mlog.Int("count", len(savedReceipts)))

	return nil
}

