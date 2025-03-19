//go:build !enterprise

package biz

func (d *OAuth2SessionUsecase) DeleteExpiredSessions() {
	return
}
