//go:build enterprise

package biz

import (
	"context"
	"fmt"
	"time"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
)

func (d *OAuth2SessionUsecase) CreateOrUpdateSession(ctx context.Context, UserUid, Sub, Sid, IdToken, RefreshToken string, deleteAfter time.Time) (uid string, err error) {
	if Sub == "" && Sid == "" {
		// sub(第三方用户标识), sid(第三方会话标识)
		// 没有第三方会话信息，不创建会话记录
		d.log.Debugf("CreateOrUpdateSession with empty sub and sid")
		return "", nil
	}
	filterBy := []pkgConst.FilterCondition{
		{Field: "sub", Operator: pkgConst.FilterOperatorEqual, Value: Sub},
		{Field: "sid", Operator: pkgConst.FilterOperatorEqual, Value: Sid},
	}
	sessions, err := d.GetSessions(ctx, filterBy)
	if err != nil {
		return "", err
	}
	if len(sessions) == 1 {
		// sub(第三方用户标识)+sid(第三方会话标识)是唯一索引，至多一条记录
		// 存在该会话记录则更新它
		return sessions[0].UID, d.SaveSession(ctx, &OAuth2Session{
			Base: Base{
				CreatedAt: sessions[0].CreatedAt,
				UpdatedAt: time.Now(),
			},
			UID:             sessions[0].UID,
			UserUID:         UserUid,
			Sub:             Sub,
			Sid:             Sid,
			IdToken:         IdToken,
			RefreshToken:    RefreshToken,
			LastLogoutEvent: "",
			DeleteAfter:     deleteAfter,
		})
	}

	// 不存在则新建会话记录
	s, err := newOAuth2Session(UserUid, Sub, Sid, IdToken, RefreshToken, deleteAfter)
	if err != nil {
		return "", fmt.Errorf("new oauth2 session failed: %v", err)
	}

	if err = d.repo.SaveSession(ctx, s); err != nil {
		return "", err
	}

	return s.UID, nil

}

func (d *OAuth2SessionUsecase) SaveSession(ctx context.Context, s *OAuth2Session) (err error) {
	if s == nil || s.UID == "" {
		return fmt.Errorf("the oauth2 session to save is nil or has no uid")
	}

	if err := d.repo.SaveSession(ctx, s); err != nil {
		return fmt.Errorf("save oauth2 session failed: %v", err)
	}

	return nil
}

func (d *OAuth2SessionUsecase) GetSessions(ctx context.Context, conditions []pkgConst.FilterCondition) (sessions []*OAuth2Session, err error) {
	if sessions, err = d.repo.GetSessions(ctx, conditions); err != nil {
		return nil, fmt.Errorf("failed to get oauth2 session: %v", err)
	}
	return sessions, nil
}

func (d *OAuth2SessionUsecase) GetSessionBySubSid(ctx context.Context, sub, sid string) (session *OAuth2Session, exist bool, err error) {
	if session, exist, err = d.repo.GetSessionBySubSid(ctx, sub, sid); err != nil {
		return session, exist, fmt.Errorf("failed to get oauth2 session: %v", err)
	}
	return session, exist, nil
}

func (d *OAuth2SessionUsecase) UpdateUserIdBySub(ctx context.Context, userid, sub string) (err error) {
	if err = d.repo.UpdateUserUidBySub(ctx, userid, sub); err != nil {
		return fmt.Errorf("failed to update oauth2 session user uid: %v", err)
	}
	return nil
}

func (d *OAuth2SessionUsecase) UpdateLogoutEvent(ctx context.Context, sub, sid, logoutIat string) (err error) {
	if err = d.repo.UpdateLogoutEvent(ctx, sub, sid, logoutIat); err != nil {
		return fmt.Errorf("failed to update oauth2 session last_logout_event: %v", err)
	}
	return nil
}

func (d *OAuth2SessionUsecase) DeleteExpiredSessions() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := d.repo.DeleteExpiredSessions(ctx); err != nil {
		d.log.Error(fmt.Sprintf("failed to delete expired oauth2 sessions: %v", err))
	}
	return
}
