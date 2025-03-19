package biz

import (
	"context"
	"time"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	pkgRand "github.com/actiontech/dms/pkg/rand"
)

type OAuth2Session struct {
	Base

	UID             string
	UserUID         string
	Sub             string
	Sid             string
	IdToken         string
	RefreshToken    string
	LastLogoutEvent string
	DeleteAfter     time.Time // 记录保留时间，在此时间之后将删除该记录
}

func newOAuth2Session(UserUID, Sub, Sid, IdToken, RefreshToken string, deleteAfter time.Time) (*OAuth2Session, error) {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}
	return &OAuth2Session{
		UID:          uid,
		UserUID:      UserUID,
		Sub:          Sub,
		Sid:          Sid,
		IdToken:      IdToken,
		RefreshToken: RefreshToken,
		DeleteAfter:  deleteAfter,
	}, nil
}

func (u *OAuth2Session) GetUID() string {
	return u.UID
}

type OAuth2SessionRepo interface {
	SaveSession(ctx context.Context, s *OAuth2Session) error
	GetSessions(ctx context.Context, conditions []pkgConst.FilterCondition) ([]*OAuth2Session, error)
	UpdateUserUidBySub(ctx context.Context, userid, sub string) error
	UpdateLogoutEvent(ctx context.Context, Sub, Sid, logoutIat string) error
	DeleteExpiredSessions(ctx context.Context) error
}

type OAuth2SessionUsecase struct {
	tx   TransactionGenerator
	repo OAuth2SessionRepo
	log  *utilLog.Helper
}

func NewOAuth2SessionUsecase(log utilLog.Logger, tx TransactionGenerator, repo OAuth2SessionRepo) *OAuth2SessionUsecase {
	return &OAuth2SessionUsecase{
		tx:   tx,
		repo: repo,
		log:  utilLog.NewHelper(log, utilLog.WithMessageKey("biz.OAuth2Session")),
	}
}
