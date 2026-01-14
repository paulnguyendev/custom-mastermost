// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"database/sql"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/request"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost/server/v8/channels/store"
	"github.com/mattermost/mattermost/server/v8/einterfaces"

	sq "github.com/mattermost/squirrel"
)

type SqlReadReceiptStore struct {
	*SqlStore
	metrics einterfaces.MetricsInterface

	selectQueryBuilder sq.SelectBuilder
}

func newSqlReadReceiptStore(sqlStore *SqlStore, metrics einterfaces.MetricsInterface) store.ReadReceiptStore {
	s := &SqlReadReceiptStore{
		SqlStore: sqlStore,
		metrics:  metrics,
	}

	s.selectQueryBuilder = s.getQueryBuilder().Select(readReceiptSliceColumns()...).From("ReadReceipts")

	return s
}

func readReceiptSliceColumns() []string {
	return []string{
		"PostId",
		"UserId",
		"ChannelId",
		"SeenAt",
		"ExpireAt",
	}
}

func (s *SqlReadReceiptStore) InvalidateReadReceiptForPostsCache(postID string) {
}

func (s *SqlReadReceiptStore) Save(rctx request.CTX, receipt *model.ReadReceipt) (*model.ReadReceipt, error) {
	receipt.PreSave()

	// Check if already exists
	existing, err := s.Get(rctx, receipt.PostID, receipt.UserID)
	if err == nil && existing != nil {
		// Already seen, return existing
		return existing, nil
	}

	// If error is not "not found", return the error
	var nfErr *store.ErrNotFound
	if err != nil && !errors.As(err, &nfErr) {
		return nil, err
	}

	query := s.getQueryBuilder().
		Insert("ReadReceipts").
		Columns(readReceiptSliceColumns()...).
		Values(
			receipt.PostID,
			receipt.UserID,
			receipt.ChannelID,
			receipt.SeenAt,
			receipt.ExpireAt,
		)

	_, err = s.GetMaster().ExecBuilder(query)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

// SaveMultiple saves multiple read receipts at once using batch insert
// It uses INSERT ... ON CONFLICT DO NOTHING to skip duplicates
func (s *SqlReadReceiptStore) SaveMultiple(rctx request.CTX, receipts []*model.ReadReceipt) ([]*model.ReadReceipt, error) {
	if len(receipts) == 0 {
		return []*model.ReadReceipt{}, nil
	}

	// PreSave all receipts
	for _, receipt := range receipts {
		receipt.PreSave()
	}

	// Build batch insert query
	query := s.getQueryBuilder().
		Insert("ReadReceipts").
		Columns(readReceiptSliceColumns()...)

	for _, receipt := range receipts {
		query = query.Values(
			receipt.PostID,
			receipt.UserID,
			receipt.ChannelID,
			receipt.SeenAt,
			receipt.ExpireAt,
		)
	}

	// Add ON CONFLICT DO NOTHING to skip duplicates
	if s.DriverName() == model.DatabaseDriverPostgres {
		query = query.Suffix("ON CONFLICT (PostId, UserId) DO NOTHING")
	} else {
		// MySQL uses ON DUPLICATE KEY UPDATE with a no-op
		query = query.Suffix("ON DUPLICATE KEY UPDATE PostId=PostId")
	}

	_, err := s.GetMaster().ExecBuilder(query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to batch save read receipts")
	}

	return receipts, nil
}

func (s *SqlReadReceiptStore) Update(rctx request.CTX, receipt *model.ReadReceipt) (*model.ReadReceipt, error) {
	query := s.getQueryBuilder().
		Update("ReadReceipts").
		Set("SeenAt", receipt.SeenAt).
		Set("ExpireAt", receipt.ExpireAt).
		Where(sq.Eq{"PostId": receipt.PostID, "UserId": receipt.UserID})

	_, err := s.GetMaster().ExecBuilder(query)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

func (s *SqlReadReceiptStore) Delete(rctx request.CTX, postID, userID string) error {
	query := s.getQueryBuilder().
		Delete("ReadReceipts").
		Where(sq.Eq{"PostId": postID, "UserId": userID})

	_, err := s.GetMaster().ExecBuilder(query)
	return err
}

func (s *SqlReadReceiptStore) DeleteByPost(rctx request.CTX, postID string) error {
	query := s.getQueryBuilder().
		Delete("ReadReceipts").
		Where(sq.Eq{"PostId": postID})

	_, err := s.GetMaster().ExecBuilder(query)
	return err
}

func (s *SqlReadReceiptStore) Get(rctx request.CTX, postID, userID string) (*model.ReadReceipt, error) {
	query := s.selectQueryBuilder.
		Where(sq.Eq{"PostId": postID, "UserId": userID})

	var receipt model.ReadReceipt
	err := s.GetReplica().GetBuilder(&receipt, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.NewErrNotFound("ReadReceipt", postID+"_"+userID)
		}

		return nil, errors.Wrapf(err, "failed to get ReadReceipt with id=%s", postID+"_"+userID)
	}

	return &receipt, nil
}

func (s *SqlReadReceiptStore) GetByPost(rctx request.CTX, postID string) ([]*model.ReadReceipt, error) {
	query := s.selectQueryBuilder.
		Where(sq.Eq{"PostId": postID})

	var receipts []*model.ReadReceipt
	err := s.GetReplica().SelectBuilder(&receipts, query)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get ReadReceipts for postId=%s", postID)
	}

	return receipts, nil
}

func (s *SqlReadReceiptStore) GetReadCountForPost(rctx request.CTX, postID string) (int64, error) {
	query := s.getQueryBuilder().
		Select("COUNT(*)").
		From("ReadReceipts").
		Where(sq.Eq{"PostId": postID})

	var count int64
	err := s.GetReplica().GetBuilder(&count, query)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *SqlReadReceiptStore) GetUnreadCountForPost(rctx request.CTX, post *model.Post) (int64, error) {
	// Count channel members who haven't read the post (excluding post author)
	// LEFT JOIN with ReadReceipts to find members without a read receipt for this post
	unreadQuery := s.getQueryBuilder().
		Select("COUNT(*)").
		From("ChannelMembers").
		LeftJoin("ReadReceipts ON ChannelMembers.UserId = ReadReceipts.UserId AND ReadReceipts.PostId = ?", post.Id).
		Where(sq.And{
			sq.Eq{"ChannelMembers.ChannelId": post.ChannelId},
			sq.NotEq{"ChannelMembers.UserId": post.UserId},
			sq.Eq{"ReadReceipts.UserId": nil},
		})

	var unreadCount int64
	// Use master to avoid stale data from replica after writing a read receipt
	err := s.GetMaster().GetBuilder(&unreadCount, unreadQuery)
	if err != nil {
		return -1, errors.Wrapf(err, "failed to get unread count for postId=%s channelId=%s", post.Id, post.ChannelId)
	}

	// Return true if no one is unread (all have read it)
	return unreadCount, nil
}

func (s *SqlReadReceiptStore) GetByPostWithUsers(rctx request.CTX, postID string, limit, offset int) ([]*model.ReadReceipt, error) {
	query := s.selectQueryBuilder.
		Where(sq.Eq{"PostId": postID}).
		OrderBy("SeenAt DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	var receipts []*model.ReadReceipt
	err := s.GetReplica().SelectBuilder(&receipts, query)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get ReadReceipts for postId=%s", postID)
	}

	return receipts, nil
}

func (s *SqlReadReceiptStore) GetForPosts(rctx request.CTX, postIDs []string) ([]*model.ReadReceipt, error) {
	if len(postIDs) == 0 {
		return []*model.ReadReceipt{}, nil
	}

	query := s.selectQueryBuilder.
		Where(sq.Eq{"PostId": postIDs})

	var receipts []*model.ReadReceipt
	err := s.GetReplica().SelectBuilder(&receipts, query)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get ReadReceipts for posts")
	}

	return receipts, nil
}
