//go:build enterprise

package biz

import (
	"context"
	"fmt"
	"time"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
)

func (d *OAuth2SessionUsecase) CreateOrUpdateSession(ctx context.Context, UserUid, Sub, Sid, IdToken, RefreshToken string) (uid string, err error) {
	filterBy := []pkgConst.FilterCondition{
		{Field: "sub", Operator: pkgConst.FilterOperatorEqual, Value: Sub},
		{Field: "sid", Operator: pkgConst.FilterOperatorEqual, Value: Sid},
	}
	sessions, err := d.GetSessions(ctx, filterBy)
	if err != nil {
		return "", err
	}
	if len(sessions) == 1 {
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
		})
	}

	s, err := newOAuth2Session(UserUid, Sub, Sid, IdToken, RefreshToken)
	if err != nil {
		return "", fmt.Errorf("new session failed: %v", err)
	}

	if err := d.repo.SaveSession(ctx, s); err != nil {
		return "", fmt.Errorf("save session failed: %v", err)
	}

	return s.UID, nil

}

func (d *OAuth2SessionUsecase) SaveSession(ctx context.Context, s *OAuth2Session) (err error) {
	if s == nil || s.UID == "" {
		return fmt.Errorf("save invalid session")
	}

	if err := d.repo.SaveSession(ctx, s); err != nil {
		return fmt.Errorf("save session failed: %v", err)
	}

	return nil
}

func (d *OAuth2SessionUsecase) GetSessions(ctx context.Context, conditions []pkgConst.FilterCondition) (sessions []*OAuth2Session, err error) {
	return d.repo.GetSessions(ctx, conditions)
}

func (d *OAuth2SessionUsecase) UpdateUserIdBySub(ctx context.Context, userid, sub string) (err error) {
	return d.repo.UpdateUserUidBySub(ctx, userid, sub)
}

func (d *OAuth2SessionUsecase) UpdateLogoutEvent(ctx context.Context, sub, sid, logoutIat string) (err error) {
	return d.repo.UpdateLogoutEvent(ctx, sub, sid, logoutIat)
}
