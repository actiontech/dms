package biz

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strconv"
	"time"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"github.com/actiontech/dms/pkg/dms-common/api/jwt"
	"github.com/actiontech/dms/pkg/dms-common/i18nPkg"
	_const "github.com/actiontech/dms/pkg/dms-common/pkg/const"
	pkgRand "github.com/actiontech/dms/pkg/rand"
	"github.com/labstack/echo/v4"

	"github.com/actiontech/dms/pkg/dms-common/pkg/aes"

	"github.com/go-ldap/ldap/v3"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type UserAuthenticationType string

const (
	UserAuthenticationTypeLDAP   UserAuthenticationType = "ldap"                  // user verify through ldap
	UserAuthenticationTypeDMS    UserAuthenticationType = _const.DmsComponentName // user verify through dms
	UserAuthenticationTypeOAUTH2 UserAuthenticationType = "oauth2"                // user verify through oauth2
)

func (u *UserAuthenticationType) String() string {
	return string(*u)
}

func ParseUserAuthenticationType(typ string) (UserAuthenticationType, error) {
	switch typ {
	case string(UserAuthenticationTypeLDAP):
		return UserAuthenticationTypeLDAP, nil
	case string(UserAuthenticationTypeDMS):
		return UserAuthenticationTypeDMS, nil
	case string(UserAuthenticationTypeOAUTH2):
		return UserAuthenticationTypeOAUTH2, nil
	default:
		return "", fmt.Errorf("invalid user authentication type: %s", typ)
	}
}

type UserStat uint

const (
	UserStatOK      UserStat = iota // 0
	UserStatDisable                 // 1

	userStatMax
)

func (u *UserStat) Uint() uint {
	return uint(*u)
}

func ParseUserStat(stat uint) (UserStat, error) {
	if stat < uint(userStatMax) {
		return UserStat(stat), nil
	}
	return 0, fmt.Errorf("invalid user stat: %d", stat)
}

type User struct {
	Base

	UID                    string
	Name                   string
	Password               string
	ThirdPartyUserID       string
	ThirdPartyUserInfo     string
	Email                  string
	Phone                  string
	WxID                   string
	Language               string
	Desc                   string
	UserAuthenticationType UserAuthenticationType
	Stat                   UserStat
	// 用户上次登录时间的时间
	LastLoginAt time.Time
	// 用户是否被删除
	Deleted bool
}

type AccessTokenInfo struct {
	UID         string
	UserID      uint
	Token       string
	ExpiredTime time.Time
}

func initUsers() []*User {
	return []*User{
		{
			UID:                    pkgConst.UIDOfUserAdmin,
			Name:                   "admin",
			Password:               "admin",
			Desc:                   "built-in admin user",
			UserAuthenticationType: UserAuthenticationTypeDMS,
			Stat:                   UserStatOK,
		},
		{
			UID:                    pkgConst.UIDOfUserSys,
			Name:                   "sys",
			Password:               "sys",
			Desc:                   "built-in sys user",
			UserAuthenticationType: UserAuthenticationTypeDMS,
			Stat:                   UserStatOK,
		},
	}
}

func newUser(args *CreateUserArgs) (*User, error) {
	if args.Name == "" {
		return nil, fmt.Errorf("name is empty")
	}
	if !args.IsDisabled {
		if args.Password == "" {
			return nil, fmt.Errorf("password is empty")
		}
	}
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}
	if args.UserAuthenticationType == "" {
		args.UserAuthenticationType = UserAuthenticationTypeDMS
	}
	return &User{
		UID:                    uid,
		Name:                   args.Name,
		Password:               args.Password,
		Email:                  args.Email,
		Phone:                  args.Phone,
		WxID:                   args.WxID,
		Desc:                   args.Desc,
		UserAuthenticationType: args.UserAuthenticationType,
		ThirdPartyUserID:       args.ThirdPartyUserID,
		ThirdPartyUserInfo:     args.ThirdPartyUserInfo,
		Stat:                   UserStatOK,
	}, nil
}

func (u *User) GetUID() string {
	return u.UID
}

type ListUsersOption struct {
	PageNumber   uint32
	LimitPerPage uint32
	OrderBy      UserField
	FilterBy     []pkgConst.FilterCondition
}

type UserRepo interface {
	SaveUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, u *User) error
	CheckUserExist(ctx context.Context, userUids []string) (exists bool, err error)
	ListUsers(ctx context.Context, opt *ListUsersOption) (users []*User, total int64, err error)
	CountUsers(ctx context.Context, opts []pkgConst.FilterCondition) (total int64, err error)
	DelUser(ctx context.Context, UserUid string) error
	GetUser(ctx context.Context, UserUid string) (*User, error)
	GetUserByName(ctx context.Context, userName string) (*User, error)
	AddUserToUserGroup(ctx context.Context, userGroupUid string, userUid string) error
	DelUserFromAllUserGroups(ctx context.Context, userUid string) error
	ReplaceUserGroupsInUser(ctx context.Context, userUid string, userGroupUids []string) error
	ReplaceOpPermissionsInUser(ctx context.Context, userUid string, OpPermissionUids []string) error
	GetUserGroupsByUser(ctx context.Context, userUid string) ([]*UserGroup, error)
	GetOpPermissionsByUser(ctx context.Context, userUid string) ([]*OpPermission, error)
	GetUserByThirdPartyUserID(ctx context.Context, thirdPartyUserUID string) (*User, error)
	SaveAccessToken(ctx context.Context, accessTokenInfo *AccessTokenInfo) error
	GetAccessTokenByUser(ctx context.Context, UserUid string) (*AccessTokenInfo, error)
}

type UserUsecase struct {
	tx                        TransactionGenerator
	repo                      UserRepo
	userGroupRepo             UserGroupRepo
	pluginUsecase             *PluginUsecase
	opPermissionUsecase       *OpPermissionUsecase
	OpPermissionVerifyUsecase *OpPermissionVerifyUsecase
	ldapConfigurationUsecase  *LDAPConfigurationUsecase
	log                       *utilLog.Helper
}

func NewUserUsecase(log utilLog.Logger, tx TransactionGenerator, repo UserRepo, userGroupRepo UserGroupRepo, pluginUsecase *PluginUsecase, opPermissionUsecase *OpPermissionUsecase,
	OpPermissionVerifyUsecase *OpPermissionVerifyUsecase, ldapConfigurationUsecase *LDAPConfigurationUsecase) *UserUsecase {
	return &UserUsecase{
		tx:                        tx,
		repo:                      repo,
		userGroupRepo:             userGroupRepo,
		pluginUsecase:             pluginUsecase,
		opPermissionUsecase:       opPermissionUsecase,
		OpPermissionVerifyUsecase: OpPermissionVerifyUsecase,
		ldapConfigurationUsecase:  ldapConfigurationUsecase,
		log:                       utilLog.NewHelper(log, utilLog.WithMessageKey("biz.user")),
	}
}

func (d *UserUsecase) UserLogin(ctx context.Context, name string, password string) (uid string, err error) {
	loginVerifier, err := d.GetUserLoginVerifier(ctx, name)
	if err != nil {
		return "", fmt.Errorf("get user login verifier failed: %v", err)
	}
	userUid, err := loginVerifier.Verify(ctx, name, password)
	if err != nil {
		return "", fmt.Errorf("verify user login failed: %v", err)
	}

	return userUid, nil
}

// GetUserLoginVerifier get login Verifier by user name and init login verifier
func (d *UserUsecase) GetUserLoginVerifier(ctx context.Context, name string) (UserLoginVerifier, error) {
	user, err := d.repo.GetUserByName(ctx, name)
	if nil != err && !errors.Is(err, pkgErr.ErrStorageNoData) {
		return nil, fmt.Errorf("get user by name error: %v", err)
	}

	ldapC, _, err := d.ldapConfigurationUsecase.GetLDAPConfiguration(ctx)
	if err != nil {
		return nil, fmt.Errorf("get ldap configuration failed: %v", err)
	}

	loginVerifierType, exist := d.getLoginVerifierType(user, ldapC)
	if err != nil {
		return nil, fmt.Errorf("get login verifier type failed: %v", err)
	}

	var userLoginVerifier UserLoginVerifier
	{

		switch loginVerifierType {
		case loginVerifierTypeLDAP:
			userLoginVerifier = &LoginLdap{
				LoginBase: LoginBase{
					user:        user,
					userUsecase: d,
				},
				userExist: exist,
				config:    ldapC,
			}
		case loginVerifierTypeDMS:
			userLoginVerifier = &LoginBase{
				user:        user,
				userUsecase: d,
			}
		case loginVerifierTypeUnknown:
			return nil, fmt.Errorf("the user login type is unsupported")
		default:
			return nil, fmt.Errorf("the user does not exist or the password is wrong")
		}

	}
	return userLoginVerifier, nil
}

type verifierType int

const (
	loginVerifierTypeUnknown verifierType = iota
	loginVerifierTypeDMS
	loginVerifierTypeLDAP
)

// determine whether the login conditions are met according to the order of login priority
func (d *UserUsecase) getLoginVerifierType(user *User, ldapC *LDAPConfiguration) (verifyType verifierType, userExist bool) {

	// ldap login condition
	if ldapC != nil && ldapC.Enable {
		if user != nil && user.UserAuthenticationType == UserAuthenticationTypeLDAP {
			return loginVerifierTypeLDAP, true
		}
		if user == nil {
			return loginVerifierTypeLDAP, false
		}
	}

	// login condition, oauth 2 and other login types of users can also log in through the account and password
	if user != nil && (user.UserAuthenticationType != UserAuthenticationTypeLDAP) {
		return loginVerifierTypeDMS, true
	}

	// no alternative login method
	return loginVerifierTypeUnknown, user != nil
}

func (d *UserUsecase) GetUserFingerprint(user *User) string {
	return fmt.Sprintf("%s_%s", user.UID, aes.Md5(user.Password))
}

type UserLoginVerifier interface {
	Verify(c context.Context, userName, password string) (userUID string, err error)
}

type LoginBase struct {
	user        *User
	userUsecase *UserUsecase
}

func (l *LoginBase) Verify(c context.Context, userName, password string) (userUID string, err error) {
	if l.user.Stat == UserStatDisable {
		return l.user.UID, fmt.Errorf("user %s not exist or can not login", userName)
	} else if l.user.Password != password {
		return l.user.UID, fmt.Errorf("user %s password not match", userName)
	}
	return l.user.UID, nil
}

type LoginLdap struct {
	LoginBase
	userExist bool
	config    *LDAPConfiguration
}

var errLdapLoginFailed = errors.New("ldap login failed, username and password do not match")

const ldapServerErrorFormat = "search user on ldap server failed: %v"

func (l *LoginLdap) Verify(ctx context.Context, userName, password string) (userUID string, err error) {

	ldapC := l.config
	var conn *ldap.Conn
	if l.config.EnableSSL {
		url := fmt.Sprintf("ldaps://%s:%s", ldapC.Host, ldapC.Port)
		conn, err = ldap.DialURL(url, ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	} else {
		url := fmt.Sprintf("ldap://%s:%s", ldapC.Host, ldapC.Port)
		conn, err = ldap.DialURL(url)
	}
	if err != nil {
		return "", fmt.Errorf("get ldap server connect failed: %v", err)
	}

	defer conn.Close()

	if err = conn.Bind(ldapC.ConnectDn, ldapC.ConnectPassword); err != nil {
		return "", fmt.Errorf("bind ldap manager user failed: %v", err)
	}
	searchRequest := ldap.NewSearchRequest(
		ldapC.BaseDn,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(%s=%s)", ldapC.UserNameRdnKey, userName),
		[]string{},
		nil,
	)
	result, err := conn.Search(searchRequest)
	if err != nil {
		return "", fmt.Errorf(ldapServerErrorFormat, err)
	}
	if len(result.Entries) == 0 {
		return "", errLdapLoginFailed
	}
	if len(result.Entries) != 1 {
		return "", fmt.Errorf(ldapServerErrorFormat, "the queried user is not unique, please check whether the relevant configuration is correct")
	}
	userDn := result.Entries[0].DN
	if err = conn.Bind(userDn, password); err != nil {
		return "", errLdapLoginFailed
	}

	// ldap bind user
	{
		// create user: ldap login without bind user
		if !l.userExist {
			userUid, err := l.userUsecase.CreateUser(ctx, pkgConst.UIDOfUserSys, &CreateUserArgs{
				Name:                   userName,
				Password:               password,
				Email:                  result.Entries[0].GetAttributeValue(ldapC.UserEmailRdnKey),
				IsDisabled:             false,
				UserAuthenticationType: UserAuthenticationTypeLDAP,
			})
			if err != nil {
				return "", err
			}
			return userUid, nil
		} else if l.user.Password != password {
			// update user: ldap login with bind user but password is different
			l.user.Password = password
			err := l.userUsecase.SaveUser(ctx, l.user)
			if err != nil {
				return "", err
			}
		}
	}

	if l.user.UID == pkgConst.UIDOfUserSys {
		return "", fmt.Errorf("sys user can not login")
	}
	if l.user.Stat != UserStatOK {
		return "", fmt.Errorf("user stat disabled")
	}

	return l.user.GetUID(), nil
}

func (d *UserUsecase) AfterUserLogin(ctx context.Context, uid string) (err error) {
	user, err := d.GetUser(ctx, uid)
	if nil != err {
		return fmt.Errorf("get user error: %v", err)
	}
	user.LastLoginAt = time.Now()
	if err := d.repo.UpdateUser(ctx, user); nil != err {
		return fmt.Errorf("update user error: %v", err)
	}
	return nil
}

type CreateUserArgs struct {
	Name                   string
	Password               string
	ThirdPartyUserID       string
	ThirdPartyUserInfo     string
	Email                  string
	Phone                  string
	WxID                   string
	Desc                   string
	UserGroupUIDs          []string
	IsDisabled             bool
	OpPermissionUIDs       []string
	UserAuthenticationType UserAuthenticationType
}

func (d *UserUsecase) CreateUser(ctx context.Context, currentUserUid string, args *CreateUserArgs) (uid string, err error) {
	// check
	{
		if isAdmin, err := d.OpPermissionVerifyUsecase.IsUserDMSAdmin(ctx, currentUserUid); err != nil {
			return "", fmt.Errorf("check user is admin failed: %v", err)
		} else if !isAdmin {
			return "", fmt.Errorf("user is not admin")
		}

		user, err := d.repo.GetUserByName(ctx, args.Name)
		if err == nil {
			return "", fmt.Errorf("user %v is exist", user.Name)
		}
		if nil != err && !errors.Is(err, pkgErr.ErrStorageNoData) {
			return "", fmt.Errorf("get user by name error: %v", user)
		}
	}

	u, err := newUser(args)
	if err != nil {
		return "", fmt.Errorf("new user failed: %v", err)
	}

	tx := d.tx.BeginTX(ctx)
	defer func() {
		if err != nil {
			err = tx.RollbackWithError(d.log, err)
		}
	}()

	if err := d.repo.SaveUser(tx, u); err != nil {
		return "", fmt.Errorf("save user failed: %v", err)
	}

	if err := d.InsureUserToUserGroups(tx, args.UserGroupUIDs, u.UID); err != nil {
		return "", fmt.Errorf("insure user to user groups failed: %v", err)
	}

	if err := d.InsureOpPermissionsInUser(tx, args.OpPermissionUIDs, u.UID); err != nil {
		return "", fmt.Errorf("insure op permissions in user failed: %v", err)
	}

	if err := tx.Commit(d.log); err != nil {
		return "", fmt.Errorf("commit tx failed: %v", err)
	}

	return u.UID, nil
}

func (d *UserUsecase) InitUsers(ctx context.Context) (err error) {
	for _, u := range initUsers() {

		_, err := d.repo.GetUser(ctx, u.GetUID())
		// already exist
		if err == nil {
			continue
		}

		// error, return directly
		if !errors.Is(err, pkgErr.ErrStorageNoData) {
			return fmt.Errorf("failed to get user permission: %w", err)
		}

		// not exist, then create it
		if err := d.repo.SaveUser(ctx, u); err != nil {
			return fmt.Errorf("failed to init user permission: %w", err)
		}

	}
	d.log.Debug("init users success")
	return nil
}

func (d *UserUsecase) CheckUserExist(ctx context.Context, userUids []string) (exists bool, err error) {
	return d.repo.CheckUserExist(ctx, userUids)
}

// InsureUserToUserGroups 确保用户属于指定的多个用户组
func (d *UserUsecase) InsureUserToUserGroups(ctx context.Context, userGroupUids []string, userUid string) (err error) {
	// check
	{
		if exist, err := d.userGroupRepo.CheckUserGroupExist(ctx, userGroupUids); err != nil {
			return fmt.Errorf("check user group exist failed: %v", err)
		} else if !exist {
			return fmt.Errorf("user group not exist")
		}
	}

	if err := d.repo.ReplaceUserGroupsInUser(ctx, userUid, userGroupUids); err != nil {
		return fmt.Errorf("replace user groups in user failed: %v", err)
	}

	return nil
}

// InsureOpPermissionsInUser 确保用户拥有指定的多个操作权限
func (d *UserUsecase) InsureOpPermissionsInUser(ctx context.Context, opPermissionUids []string, userUid string) (err error) {
	// check
	{
		// 检查是否全局操作权限，因为用户为全局类型，只能绑定全局操作权限
		if isGlobal, err := d.opPermissionUsecase.IsGlobalOpPermissions(ctx, opPermissionUids); err != nil {
			return fmt.Errorf("check is global op permissions failed: %v", err)
		} else if !isGlobal {
			return fmt.Errorf("op permissions must be global")
		}

	}

	if err := d.repo.ReplaceOpPermissionsInUser(ctx, userUid, opPermissionUids); err != nil {
		return fmt.Errorf("replace op permissions in user failed: %v", err)
	}

	return nil
}

func (d *UserUsecase) ListUser(ctx context.Context, option *ListUsersOption) (users []*User, total int64, err error) {
	users, total, err = d.repo.ListUsers(ctx, option)
	if err != nil {
		return nil, 0, fmt.Errorf("list users failed: %v", err)
	}
	return users, total, nil
}

func (d *UserUsecase) DelUser(ctx context.Context, currentUserUid, UserUid string) (err error) {
	// check
	{
		if UserUid == pkgConst.UIDOfUserAdmin || UserUid == pkgConst.UIDOfUserSys {
			return fmt.Errorf("can not delete user admin or sys")
		}
		if isAdmin, err := d.OpPermissionVerifyUsecase.IsUserDMSAdmin(ctx, currentUserUid); err != nil {
			return fmt.Errorf("check user is admin failed: %v", err)
		} else if !isAdmin {
			return fmt.Errorf("user is not admin")
		}
	}

	ds, err := d.repo.GetUser(ctx, UserUid)
	if err != nil {
		return fmt.Errorf("get user failed: %v", err)
	}

	// 调用其他服务对用户进行预检查
	if err := d.pluginUsecase.DelUserPreCheck(ctx, ds.GetUID()); err != nil {
		return fmt.Errorf("precheck del user failed: %v", err)
	}

	tx := d.tx.BeginTX(ctx)
	defer func() {
		if err != nil {
			err = tx.RollbackWithError(d.log, err)
		}
	}()

	if err := d.repo.DelUserFromAllUserGroups(tx, UserUid); err != nil {
		return fmt.Errorf("del user from all user groups failed: %v", err)
	}

	if err := d.repo.DelUser(tx, UserUid); nil != err {
		return fmt.Errorf("delete user error: %v", err)
	}

	if err := tx.Commit(d.log); err != nil {
		return fmt.Errorf("commit tx failed: %v", err)
	}

	return nil
}

func (d *UserUsecase) GetUserGroups(ctx context.Context, userUid string) (userGroups []*UserGroup, err error) {
	userGroups, err = d.repo.GetUserGroupsByUser(ctx, userUid)
	if err != nil {
		return nil, fmt.Errorf("get user groups by user failed: %v", err)
	}
	return userGroups, nil
}

func (d *UserUsecase) GetUserOpPermissions(ctx context.Context, userUid string) (ops []*OpPermission, err error) {
	ops, err = d.repo.GetOpPermissionsByUser(ctx, userUid)
	if err != nil {
		return nil, fmt.Errorf("get op permissions by user failed: %v", err)
	}
	return ops, nil
}

func (d *UserUsecase) GetUser(ctx context.Context, userUid string) (*User, error) {
	return d.repo.GetUser(ctx, userUid)
}

func (d *UserUsecase) UpdateUser(ctx context.Context, currentUserUid, updateUserUid string, isDisabled bool,
	password, email, phone, wxId, language *string, userGroupUids []string, opPermissionUids []string) error {
	// checks
	{
		if isDisabled {
			if currentUserUid == updateUserUid {
				return fmt.Errorf("can not disable current user")
			}
			if pkgConst.UIDOfUserAdmin == updateUserUid {
				return fmt.Errorf("can not disable admin user")
			}
			if pkgConst.UIDOfUserSys == updateUserUid {
				return fmt.Errorf("can not disable sys user")
			}
		}

		if isAdmin, err := d.OpPermissionVerifyUsecase.IsUserDMSAdmin(ctx, currentUserUid); err != nil {
			return fmt.Errorf("check user is admin failed: %v", err)
		} else if !isAdmin {
			return fmt.Errorf("user is not admin")
		}
	}

	user, err := d.GetUser(ctx, updateUserUid)
	if err != nil {
		return fmt.Errorf("get user failed: %v", err)
	}

	if isDisabled {
		user.Stat = UserStatDisable
	} else {
		user.Stat = UserStatOK
	}

	if password != nil {
		user.Password = *password
	}
	if email != nil {
		user.Email = *email
	}
	if phone != nil {
		user.Phone = *phone
	}
	if wxId != nil {
		user.WxID = *wxId
	}
	if language != nil {
		user.Language = *language
	}

	if user.Stat == UserStatOK && user.Password == "" {
		return fmt.Errorf("password is needed when user is enabled")
	}

	tx := d.tx.BeginTX(ctx)
	defer func() {
		if err != nil {
			err = tx.RollbackWithError(d.log, err)
		}
	}()

	if err := d.InsureUserToUserGroups(tx, userGroupUids, user.GetUID()); err != nil {
		return fmt.Errorf("insure user to user groups failed: %v", err)
	}

	if err := d.InsureOpPermissionsInUser(tx, opPermissionUids, user.GetUID()); err != nil {
		return fmt.Errorf("insure op permissions in user failed: %v", err)
	}

	if err := d.repo.UpdateUser(tx, user); nil != err {
		return fmt.Errorf("update user error: %v", err)
	}

	if err := tx.Commit(d.log); err != nil {
		return fmt.Errorf("commit tx failed: %v", err)
	}
	return nil
}

func (d *UserUsecase) UpdateCurrentUser(ctx context.Context, currentUserUid string, oldPassword, password, email, phone, wxId, language *string) error {
	user, err := d.GetUser(ctx, currentUserUid)
	if err != nil {
		return fmt.Errorf("get user failed: %v", err)
	}

	// update password
	if oldPassword != nil && password != nil {
		if user.UserAuthenticationType == UserAuthenticationTypeLDAP {
			return fmt.Errorf("the password of the ldap user cannot be changed or reset, because this password is meaningless")
		}
		if user.Password != *oldPassword {
			return fmt.Errorf("old password is wrong")
		}
		user.Password = *password
	}

	if email != nil {
		user.Email = *email
	}
	if phone != nil {
		user.Phone = *phone
	}
	if wxId != nil {
		user.WxID = *wxId
	}
	if language != nil {
		user.Language = *language
	}

	if err := d.repo.UpdateUser(ctx, user); nil != err {
		return fmt.Errorf("update current user error: %v", err)
	}

	return nil
}

func (d *UserUsecase) GetUserByThirdPartyUserID(ctx context.Context, userUid string) (*User, bool, error) {
	user, err := d.repo.GetUserByThirdPartyUserID(ctx, userUid)
	if err != nil {
		if errors.Is(err, pkgErr.ErrStorageNoData) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return user, true, nil
}

func (d *UserUsecase) GetUserByName(ctx context.Context, userName string) (*User, bool, error) {
	user, err := d.repo.GetUserByName(ctx, userName)
	if err != nil {
		if errors.Is(err, pkgErr.ErrStorageNoData) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return user, true, nil
}

func (d *UserUsecase) SaveUser(ctx context.Context, user *User) error {
	return d.repo.UpdateUser(ctx, user)
}

func (d *UserUsecase) GetBizUserWithNameByUids(ctx context.Context, uids []string) []UIdWithName {
	if len(uids) == 0 {
		return []UIdWithName{}
	}
	uidWithNameCacheCache.ulock.Lock()
	defer uidWithNameCacheCache.ulock.Unlock()
	if uidWithNameCacheCache.UserCache == nil {
		uidWithNameCacheCache.UserCache = make(map[string]UIdWithName)
	}
	ret := make([]UIdWithName, 0)
	for _, uid := range uids {
		userCache, ok := uidWithNameCacheCache.UserCache[uid]
		if !ok {
			userCache = UIdWithName{
				Uid: uid,
			}
			user, err := d.repo.GetUser(ctx, uid)
			if err == nil {
				userCache.Name = user.Name
				uidWithNameCacheCache.UserCache[user.UID] = userCache
			} else {
				d.log.Errorf("get user for cache err: %v", err)
			}
		}
		ret = append(ret, userCache)
	}
	return ret
}

func (d *UserUsecase) SaveAccessToken(ctx context.Context, userId string, token string, expiredTime time.Time) error {
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		return err
	}
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return err
	}

	tokenInfo := &AccessTokenInfo{UID: uid, UserID: uint(userIdInt), Token: token, ExpiredTime: expiredTime}
	return d.repo.SaveAccessToken(ctx, tokenInfo)
}

func (d *UserUsecase) GetAccessTokenByUser(ctx context.Context, UserUid string) (*AccessTokenInfo, error) {
	accessTokenInfo, err := d.repo.GetAccessTokenByUser(ctx, UserUid)
	if err != nil {
		return nil, err
	}
	return accessTokenInfo, nil
}

func (d *UserUsecase) GetUserLanguageByEchoCtx(c echo.Context) string {
	uid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return ""
	}
	if uid == pkgConst.UIDOfUserSys {
		// 系统用户直接通过请求头AcceptLanguage确定语言
		return i18nPkg.GetLangByAcceptLanguage(c)
	}
	user, err := d.GetUser(c.Request().Context(), uid)
	if err != nil {
		return ""
	}
	return user.Language
}
