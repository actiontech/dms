package storage

import (
	"context"
	"fmt"

	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"github.com/actiontech/dms/internal/dms/storage/model"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"

	"gorm.io/gorm"
)

var _ biz.SessionRepo = (*SessionRepo)(nil)

type SessionRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewSessionRepo(log utilLog.Logger, s *Storage) *SessionRepo {
	return &SessionRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.Session"))}
}

func (d *SessionRepo) SaveSession(ctx context.Context, s *biz.Session) error {
	session, err := convertBizSession(s)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz Session: %v", err))
	}

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Save(session).Error; err != nil {
			return fmt.Errorf("failed to save Session: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *SessionRepo) GetSessions(ctx context.Context, conditions []pkgConst.FilterCondition) (sessions []*biz.Session, err error) {
	var results []*model.Session
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		// find models
		{
			db := tx.WithContext(ctx)
			for _, f := range conditions {
				db = gormWhere(db, f)
			}
			db = db.Find(&results)
			if err := db.Error; err != nil {
				return fmt.Errorf("failed to list db service: %v", err)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// convert model to biz
	for _, res := range results {
		ds, err := convertModelSession(res)
		if err != nil {
			return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model session: %w", err))
		}
		sessions = append(sessions, ds)
	}
	return sessions, nil
}

func (d *SessionRepo) UpdateUserUidBySub(ctx context.Context, userUid string, oauth2Sub string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.Session{}).Where("oauth2_sub = ?", oauth2Sub).Update("user_uid", userUid).Error; err != nil {
			return fmt.Errorf("failed to update session: %v", err)
		}
		return nil
	})
}

func (d *SessionRepo) UpdateLogoutEvent(ctx context.Context, oauth2Sub, oauth2Sid, logoutIat string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.Session{}).
			Where("oauth2_sub = ?", oauth2Sub).
			Where("oauth2_sid = ?", oauth2Sid).
			Update("oauth2_last_logout_event", logoutIat).Error; err != nil {
			return fmt.Errorf("failed to update session oauth2_last_logout_event: %v", err)
		}
		return nil
	})
}
