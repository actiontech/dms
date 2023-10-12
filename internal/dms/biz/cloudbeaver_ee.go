//go:build enterprise

package biz

import "context"

// A provision DBAccount
type TempDBAccount struct {
	// the dbaccount user
	User string `json:"user"`
	// the dbaccount password
	Password string `json:"password"`
	// the datasource's uid
	DbServiceUid string `json:"db_service_uid"`
}

func ResetDbServiceByAuth(ctx context.Context, activeDBServices []*DBService) ([]*DBService, error) {
	dbaccounts, err := ListAuthDbAccount(ctx)
	if err != nil {
		return nil, err
	}
	ret := make([]*DBService, 0)
	for _, activeDBService := range activeDBServices {
		for _, dbaccount := range dbaccounts {
			// use db account instead of admin account
			if dbaccount.DbServiceUid == activeDBService.UID {
				activeDBService.User = dbaccount.User
				activeDBService.Password = dbaccount.Password
				ret = append(ret, activeDBService)
			}
		}
	}

	return ret, nil
}

func ListAuthDbAccount(ctx context.Context) ([]TempDBAccount, error) {
	// http : list db account
	// todo 
	return []TempDBAccount{
	}, nil
}
