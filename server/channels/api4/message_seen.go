// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
)

func (api *API) InitMessageSeen() {
	api.BaseRoutes.PostForUser.Handle("/seen", api.APISessionRequired(markMessageAsSeen)).Methods(http.MethodPost)
	api.BaseRoutes.Post.Handle("/seen", api.APISessionRequired(getSeenUsersForPost)).Methods(http.MethodGet)
	api.BaseRoutes.Post.Handle("/seen/count", api.APISessionRequired(getSeenCountForPost)).Methods(http.MethodGet)
}

func markMessageAsSeen(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequirePostId().RequireUserId()
	if c.Err != nil {
		return
	}

	if c.Params.UserId != c.AppContext.Session().UserId {
		c.SetPermissionError(model.PermissionEditOtherUsers)
		return
	}

	if !c.App.SessionHasPermissionToChannelByPost(*c.AppContext.Session(), c.Params.PostId, model.PermissionReadChannel) {
		c.SetPermissionError(model.PermissionReadChannel)
		return
	}

	receipt, err := c.App.MarkMessageAsSeen(c.AppContext, c.Params.PostId, c.Params.UserId)
	if err != nil {
		c.Err = err
		return
	}

	if receipt == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := json.NewEncoder(w).Encode(receipt); err != nil {
		c.Logger.Warn("Error while writing response", mlog.Err(err))
	}
}

func getSeenUsersForPost(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequirePostId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionToChannelByPost(*c.AppContext.Session(), c.Params.PostId, model.PermissionReadChannel) {
		c.SetPermissionError(model.PermissionReadChannel)
		return
	}

	// Only allow post author to view seen users
	post, err := c.App.GetSinglePost(c.AppContext, c.Params.PostId, false)
	if err != nil {
		c.Err = err
		return
	}

	if post.UserId != c.AppContext.Session().UserId {
		c.SetPermissionError(model.PermissionReadChannel)
		return
	}

	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	receipts, err := c.App.GetSeenUsersForPost(c.AppContext, c.Params.PostId, limit, offset)
	if err != nil {
		c.Err = err
		return
	}

	if err := json.NewEncoder(w).Encode(receipts); err != nil {
		c.Logger.Warn("Error while writing response", mlog.Err(err))
	}
}

func getSeenCountForPost(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequirePostId()
	if c.Err != nil {
		return
	}

	if !c.App.SessionHasPermissionToChannelByPost(*c.AppContext.Session(), c.Params.PostId, model.PermissionReadChannel) {
		c.SetPermissionError(model.PermissionReadChannel)
		return
	}

	// Only allow post author to view seen count
	post, err := c.App.GetSinglePost(c.AppContext, c.Params.PostId, false)
	if err != nil {
		c.Err = err
		return
	}

	if post.UserId != c.AppContext.Session().UserId {
		c.SetPermissionError(model.PermissionReadChannel)
		return
	}

	count, err := c.App.GetSeenCountForPost(c.AppContext, c.Params.PostId)
	if err != nil {
		c.Err = err
		return
	}

	response := map[string]int64{"count": count}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		c.Logger.Warn("Error while writing response", mlog.Err(err))
	}
}

