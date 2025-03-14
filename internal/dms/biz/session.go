package biz

import (
	"context"
	"fmt"
	"time"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	pkgRand "github.com/actiontech/dms/pkg/rand"
)

type Session struct {
	Base

	UID                   string
	UserUID               string
	OAuth2Sub             string
	OAuth2Sid             string
	OAuth2IdToken         string
	OAuth2RefreshToken    string
	OAuth2LastLogoutEvent string
}

func newSession(UserUID, OAuth2Sub, OAuth2Sid, OAuth2IdToken, OAuth2RefreshToken string) (*Session, error) {
	if UserUID == "" {
		return nil, fmt.Errorf("userId is empty")
	}

	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}
	return &Session{
		UID:                uid,
		UserUID:            UserUID,
		OAuth2Sub:          OAuth2Sub,
		OAuth2Sid:          OAuth2Sid,
		OAuth2IdToken:      OAuth2IdToken,
		OAuth2RefreshToken: OAuth2RefreshToken,
	}, nil
}

func (u *Session) GetUID() string {
	return u.UID
}

type SessionRepo interface {
	SaveSession(ctx context.Context, s *Session) error
	GetSessions(ctx context.Context, conditions []pkgConst.FilterCondition) ([]*Session, error)
	UpdateUserUidBySub(ctx context.Context, userid, sub string) error
	UpdateLogoutEvent(ctx context.Context, oauth2Sub, oauth2Sid, logoutIat string) error
}

type SessionUsecase struct {
	tx   TransactionGenerator
	repo SessionRepo
	log  *utilLog.Helper
}

func NewSessionUsecase(log utilLog.Logger, tx TransactionGenerator, repo SessionRepo) *SessionUsecase {
	return &SessionUsecase{
		tx:   tx,
		repo: repo,
		log:  utilLog.NewHelper(log, utilLog.WithMessageKey("biz.Session")),
	}
}

func (d *SessionUsecase) CreateOrUpdateSession(ctx context.Context, UserUid, OAuth2Sub, OAuth2Sid, OAuth2IdToken, OAuth2RefreshToken string) (uid string, err error) {
	filterBy := []pkgConst.FilterCondition{
		{Field: "oauth2_sub", Operator: pkgConst.FilterOperatorEqual, Value: OAuth2Sub},
		{Field: "oauth2_sid", Operator: pkgConst.FilterOperatorEqual, Value: OAuth2Sid},
	}
	sessions, err := d.GetSessions(ctx, filterBy)
	if err != nil {
		return "", err
	}
	if len(sessions) == 1 {
		return sessions[0].UID, d.SaveSession(ctx, &Session{
			Base: Base{
				CreatedAt: sessions[0].CreatedAt,
				UpdatedAt: time.Now(),
			},
			UID:                   sessions[0].UID,
			UserUID:               UserUid,
			OAuth2Sub:             OAuth2Sub,
			OAuth2Sid:             OAuth2Sid,
			OAuth2IdToken:         OAuth2IdToken,
			OAuth2RefreshToken:    OAuth2RefreshToken,
			OAuth2LastLogoutEvent: "",
		})
	}

	s, err := newSession(UserUid, OAuth2Sub, OAuth2Sid, OAuth2IdToken, OAuth2RefreshToken)
	if err != nil {
		return "", fmt.Errorf("new session failed: %v", err)
	}

	if err := d.repo.SaveSession(ctx, s); err != nil {
		return "", fmt.Errorf("save session failed: %v", err)
	}

	return s.UID, nil
}

func (d *SessionUsecase) SaveSession(ctx context.Context, s *Session) (err error) {
	if s == nil || s.UID == "" {
		return fmt.Errorf("save invalid session")
	}

	if err := d.repo.SaveSession(ctx, s); err != nil {
		return fmt.Errorf("save session failed: %v", err)
	}

	return nil
}

func (d *SessionUsecase) GetSessions(ctx context.Context, conditions []pkgConst.FilterCondition) (sessions []*Session, err error) {
	return d.repo.GetSessions(ctx, conditions)
}

func (d *SessionUsecase) UpdateUserIdBySub(ctx context.Context, userid, sub string) (err error) {
	return d.repo.UpdateUserUidBySub(ctx, userid, sub)
}

func (d *SessionUsecase) UpdateLogoutEvent(ctx context.Context, oauth2Sub, oauth2Sid, logoutIat string) (err error) {
	return d.repo.UpdateLogoutEvent(ctx, oauth2Sub, oauth2Sid, logoutIat)
}
