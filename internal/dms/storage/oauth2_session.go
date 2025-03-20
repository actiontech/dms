package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"github.com/actiontech/dms/internal/dms/storage/model"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"

	"gorm.io/gorm"
)

var _ biz.OAuth2SessionRepo = (*OAuth2SessionRepo)(nil)

type OAuth2SessionRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewOAuth2SessionRepo(log utilLog.Logger, s *Storage) *OAuth2SessionRepo {
	return &OAuth2SessionRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.OAuth2Session"))}
}

func (d *OAuth2SessionRepo) SaveSession(ctx context.Context, s *biz.OAuth2Session) error {
	session, err := convertBizOAuth2Session(s)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz oauth2 session: %v", err))
	}

	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		return tx.WithContext(ctx).Save(session).Error
	})
}

func (d *OAuth2SessionRepo) GetSessions(ctx context.Context, conditions []pkgConst.FilterCondition) (sessions []*biz.OAuth2Session, err error) {
	var results []*model.OAuth2Session
	err = transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		db := tx.WithContext(ctx)
		for _, f := range conditions {
			db = gormWhere(db, f)
		}
		return db.Find(&results).Error
	})
	if err != nil {
		return nil, err
	}

	// convert model to biz
	for _, res := range results {
		ds, err := convertModelOAuth2Session(res)
		if err != nil {
			return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model oauth2 session: %w", err))
		}
		sessions = append(sessions, ds)
	}
	return sessions, nil
}

func (d *OAuth2SessionRepo) GetSessionBySubSid(ctx context.Context, sub, sid string) (session *biz.OAuth2Session, exist bool, err error) {
	record := &model.OAuth2Session{}
	err = d.db.WithContext(ctx).Where("sub = ? and sid = ?", sub, sid).First(record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}

	// convert model to biz
	bizSession, err := convertModelOAuth2Session(record)
	if err != nil {
		return nil, true, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model oauth2 session: %w", err))
	}

	return bizSession, true, nil
}

func (d *OAuth2SessionRepo) UpdateUserUidBySub(ctx context.Context, userUid string, sub string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		return tx.WithContext(ctx).Model(&model.OAuth2Session{}).Where("sub = ?", sub).Update("user_uid", userUid).Error
	})
}

func (d *OAuth2SessionRepo) UpdateLogoutEvent(ctx context.Context, sub, sid, logoutIat string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		return tx.WithContext(ctx).Model(&model.OAuth2Session{}).
			Where("sub = ?", sub).
			Where("sid = ?", sid).
			Update("last_logout_event", logoutIat).Error
	})
}

func (d *OAuth2SessionRepo) DeleteExpiredSessions(ctx context.Context) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		return tx.WithContext(ctx).Unscoped().Delete(&model.OAuth2Session{}, "delete_after < ?", time.Now()).Error
	})
}
