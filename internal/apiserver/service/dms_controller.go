package service

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	aV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/apiserver/conf"
	apiError "github.com/actiontech/dms/internal/apiserver/pkg/error"
	"github.com/actiontech/dms/internal/dms/pkg/constant"
	"github.com/actiontech/dms/internal/dms/service"

	dmsV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	"github.com/actiontech/dms/pkg/dms-common/api/jwt"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"

	"github.com/labstack/echo/v4"
)

type DMSController struct {
	DMS *service.DMSService

	shutdownCallback func() error
}

func NewDMSController(logger utilLog.Logger, opts *conf.DMSOptions) (*DMSController, error) {
	dmsService, err := service.NewAndInitDMSService(logger, opts)
	if nil != err {
		return nil, fmt.Errorf("failed to init dms service: %v", err)
	}
	return &DMSController{
		// log:   log.NewHelper(log.With(logger, "module", "controller/DMS")),
		DMS: dmsService,
		shutdownCallback: func() error {
			if err := dmsService.Shutdown(); err != nil {
				return err
			}
			return nil
		},
	}, nil
}

func (a *DMSController) Shutdown() error {
	if nil != a.shutdownCallback {
		return a.shutdownCallback()
	}
	return nil
}

// swagger:route POST /v1/dms/projects/{project_uid}/db_services dms AddDBService
//
// Add DB Service.
//
//	responses:
//	  200: body:AddDBServiceReply
//	  default: body:GenericResp
func (d *DMSController) AddDBService(c echo.Context) error {
	req := new(aV1.AddDBServiceReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, err := d.DMS.AddDBService(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/db_services dms ListDBServices
//
// List db service.
//
//	responses:
//	  200: body:ListDBServiceReply
//	  default: body:GenericResp
func (d *DMSController) ListDBServices(c echo.Context) error {
	req := new(dmsV1.ListDBServiceReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	reply, err := d.DMS.ListDBServices(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/db_services/driver_options dms ListDBServiceDriverOption
//
// List db service driver option.
//
//	responses:
//	  200: body:ListDBServiceDriverOptionReply
//	  default: body:GenericResp
func (d *DMSController) ListDBServiceDriverOption(c echo.Context) error {
	reply, err := d.DMS.ListDBServiceDriverOption(c.Request().Context())
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route DELETE /v1/dms/projects/{project_uid}/db_services/{db_service_uid} dms DelDBService
//
// Delete a DB Service.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (a *DMSController) DelDBService(c echo.Context) error {
	req := &aV1.DelDBServiceReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	err = a.DMS.DelDBService(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route PUT /v1/dms/projects/{project_uid}/db_services/{db_service_uid} dms UpdateDBService
//
// update a DB Service.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (a *DMSController) UpdateDBService(c echo.Context) error {
	req := &aV1.UpdateDBServiceReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	err = a.DMS.UpdateDBService(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route POST /v1/dms/projects/{project_uid}/db_services/connection dms CheckDBServiceIsConnectable
//
// check if the db_service is connectable.
//
//	responses:
//	  200: body:CheckDBServiceIsConnectableReply
//	  default: body:GenericResp
func (d *DMSController) CheckDBServiceIsConnectable(c echo.Context) error {
	var req aV1.CheckDBServiceIsConnectableReq
	err := bindAndValidateReq(c, &req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := d.DMS.CheckDBServiceIsConnectable(c.Request().Context(), &req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route POST /v1/dms/projects/{project_uid}/db_services/{db_service_uid}/connection dms CheckDBServiceIsConnectableById
//
// check if the db_service is connectable.
//
//	responses:
//	  200: body:CheckDBServiceIsConnectableReply
//	  default: body:GenericResp
func (d *DMSController) CheckDBServiceIsConnectableById(c echo.Context) error {
	var req aV1.CheckDBServiceIsConnectableByIdReq
	err := bindAndValidateReq(c, &req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := d.DMS.CheckDBServiceIsConnectableById(c.Request().Context(), &req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/database_source_services dms ListDatabaseSourceServices
//
// List database source service.
//
//	responses:
//	  200: body:ListDatabaseSourceServicesReply
//	  default: body:GenericResp
func (d *DMSController) ListDatabaseSourceServices(c echo.Context) error {
	req := new(aV1.ListDatabaseSourceServicesReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	reply, err := d.DMS.ListDatabaseSourceService(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/database_source_services/{database_source_service_uid} dms GetDatabaseSourceService
//
// Get database source service.
//
//	responses:
//	  200: body:GetDatabaseSourceServiceReply
//	  default: body:GenericResp
func (d *DMSController) GetDatabaseSourceService(c echo.Context) error {
	req := new(aV1.GetDatabaseSourceServiceReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	reply, err := d.DMS.GetDatabaseSourceService(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route POST /v1/dms/projects/{project_uid}/database_source_services dms AddDatabaseSourceService
//
// Add database source service.
//
//	responses:
//	  200: body:AddDatabaseSourceServiceReply
//	  default: body:GenericResp
func (d *DMSController) AddDatabaseSourceService(c echo.Context) error {
	req := new(aV1.AddDatabaseSourceServiceReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, err := d.DMS.AddDatabaseSourceService(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route PUT /v1/dms/projects/{project_uid}/database_source_services/{database_source_service_uid} dms UpdateDatabaseSourceService
//
// update database source service.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (d *DMSController) UpdateDatabaseSourceService(c echo.Context) error {
	req := &aV1.UpdateDatabaseSourceServiceReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	err = d.DMS.UpdateDatabaseSourceService(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route DELETE /v1/dms/projects/{project_uid}/database_source_services/{database_source_service_uid} dms DeleteDatabaseSourceService
//
// Delete database source service.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (d *DMSController) DeleteDatabaseSourceService(c echo.Context) error {
	req := &aV1.DeleteDatabaseSourceServiceReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	err = d.DMS.DeleteDatabaseSourceService(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route GET /v1/dms/projects/{project_uid}/database_source_services/tips dms ListDatabaseSourceServiceTips
//
// List database source service tips.
//
//	responses:
//	  200: body:ListDatabaseSourceServiceTipsReply
//	  default: body:GenericResp
func (d *DMSController) ListDatabaseSourceServiceTips(c echo.Context) error {
	reply, err := d.DMS.ListDatabaseSourceServiceTips(c.Request().Context())
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route POST /v1/dms/projects/{project_uid}/database_source_services/{database_source_service_uid}/sync dms SyncDatabaseSourceService
//
// Sync database source service.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (d *DMSController) SyncDatabaseSourceService(c echo.Context) error {
	req := &aV1.SyncDatabaseSourceServiceReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	err = d.DMS.SyncDatabaseSourceService(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkResp(c)
}

// swagger:route GET /v1/dms/basic_info dms GetBasicInfo
//
// get basic info.
//
//	responses:
//	  200: body:GetBasicInfoReply
//	  default: body:GenericResp
func (d *DMSController) GetBasicInfo(c echo.Context) error {
	reply, err := d.DMS.GetBasicInfo(c.Request().Context())
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/personalization/logo dms GetStaticLogo
//
// get logo
//
//	Produces:
//	- application/octet-stream
//
//	responses:
//	  200: GetStaticLogoReply
//	  default: body:GenericResp
func (d *DMSController) GetStaticLogo(c echo.Context) error {
	reply, contentType, err := d.DMS.GetStaticLogo(c.Request().Context())
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return c.Blob(http.StatusOK, contentType, reply.File)
}

// swagger:route POST /v1/dms/personalization dms Personalization
//
// personalize [title, logo]
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (d *DMSController) Personalization(c echo.Context) error {
	req := &aV1.PersonalizationReq{}

	fileHeader, err := c.FormFile("file")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	req.File = fileHeader

	err = bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	err = d.DMS.Personalization(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkResp(c)
}

// swagger:route POST /v1/dms/sessions dms AddSession
//
// Add a session.
//
//	responses:
//	  200: body:AddSessionReply
//	  default: body:GenericResp
func (a *DMSController) AddSession(c echo.Context) error {
	req := new(aV1.AddSessionReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := a.DMS.VerifyUserLogin(c.Request().Context(), &aV1.VerifyUserLoginReq{
		UserName: req.Session.UserName,
		Password: req.Session.Password,
	})
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	if reply.Data.VerifyFailedMsg != "" {
		return NewErrResp(c, errors.New(reply.Data.VerifyFailedMsg), apiError.BadRequestErr)
	}

	// Create token with claims
	token, err := jwt.GenJwtToken(jwt.WithUserId(reply.Data.UserUid))
	if nil != err {
		return NewErrResp(c, err, apiError.APIServerErr)
	}

	err = a.DMS.AfterUserLogin(c.Request().Context(), &aV1.AfterUserLoginReq{
		UserUid: reply.Data.UserUid,
	})
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	c.SetCookie(&http.Cookie{
		Name:    constant.DMSToken,
		Value:   token,
		Path:    "/",
		Expires: time.Now().Add(24 * time.Hour),
	})

	return NewOkRespWithReply(c, &aV1.AddSessionReply{
		Data: struct {
			Token string `json:"token"`
		}{
			Token: token,
		},
	})
}

// swagger:route GET /v1/dms/sessions/user dms GetUserBySession
//
// Get current user.
//
//	responses:
//	  200: body:GetUserBySessionReply
//	  default: body:GenericResp
func (a *DMSController) GetUserBySession(c echo.Context) error {
	uid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := a.DMS.GetCurrentUser(c.Request().Context(), &aV1.GetUserBySessionReq{UserUid: uid})
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route POST /v1/dms/users dms AddUser
//
// Add user.
//
//	responses:
//	  200: body:AddUserReply
//	  default: body:GenericResp
func (d *DMSController) AddUser(c echo.Context) error {
	req := new(aV1.AddUserReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, err := d.DMS.AddUser(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route PUT /v1/dms/users/{user_uid} dms UpdateUser
//
// Update a user.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (d *DMSController) UpdateUser(c echo.Context) error {
	req := new(aV1.UpdateUserReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = d.DMS.UpdateUser(c.Request().Context(), req, currentUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route PUT /v1/dms/users dms UpdateCurrentUser
//
// Update current user.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (d *DMSController) UpdateCurrentUser(c echo.Context) error {
	req := new(aV1.UpdateCurrentUserReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = d.DMS.UpdateCurrentUser(c.Request().Context(), req, currentUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route DELETE /v1/dms/users/{user_uid} dms DelUser
//
// Delete a user.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (a *DMSController) DelUser(c echo.Context) error {
	req := &aV1.DelUserReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = a.DMS.DelUser(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route GET /v1/dms/users dms ListUsers
//
// List users.
//
//	responses:
//	  200: body:ListUserReply
//	  default: body:GenericResp
func (d *DMSController) ListUsers(c echo.Context) error {
	req := new(dmsV1.ListUserReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := d.DMS.ListUsers(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/users/{user_uid}/op_permission dms GetUserOpPermission
//
// Get user op permission info, This API is used by other component such as sqle&auth to check user permissions.
//
//	responses:
//	  200: body:GetUserOpPermissionReply
//	  default: body:GenericResp
func (a *DMSController) GetUserOpPermission(c echo.Context) error {
	req := new(dmsV1.GetUserOpPermissionReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := a.DMS.GetUserOpPermission(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/users/{user_uid} dms GetUser
//
// Get user info, This API is used by other component such as sqle&auth to get user info.
//
//	responses:
//	  200: body:GetUserReply
//	  default: body:GenericResp
func (a *DMSController) GetUser(c echo.Context) error {
	req := new(dmsV1.GetUserReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := a.DMS.GetUser(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route POST /v1/dms/user_groups dms AddUserGroup
//
// Add user group.
//
//	responses:
//	  200: body:AddUserGroupReply
//	  default: body:GenericResp
func (d *DMSController) AddUserGroup(c echo.Context) error {
	req := new(aV1.AddUserGroupReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, err := d.DMS.AddUserGroup(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route PUT /v1/dms/user_groups/{user_group_uid} dms UpdateUserGroup
//
// Update a user group.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (d *DMSController) UpdateUserGroup(c echo.Context) error {
	req := new(aV1.UpdateUserGroupReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = d.DMS.UpdateUserGroup(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route DELETE /v1/dms/user_groups/{user_group_uid} dms DelUserGroup
//
// Delete a user group.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (a *DMSController) DelUserGroup(c echo.Context) error {
	req := &aV1.DelUserGroupReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = a.DMS.DelUserGroup(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route GET /v1/dms/user_groups dms ListUserGroups
//
// List user groups.
//
//	responses:
//	  200: body:ListUserGroupReply
//	  default: body:GenericResp
func (d *DMSController) ListUserGroups(c echo.Context) error {
	req := new(aV1.ListUserGroupReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := d.DMS.ListUserGroups(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route POST /v1/dms/roles dms AddRole
//
// Add role.
//
//	responses:
//	  200: body:AddRoleReply
//	  default: body:GenericResp
func (d *DMSController) AddRole(c echo.Context) error {
	req := new(aV1.AddRoleReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, err := d.DMS.AddRole(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route PUT /v1/dms/roles/{role_uid} dms UpdateRole
//
// Update a role.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (d *DMSController) UpdateRole(c echo.Context) error {
	req := new(aV1.UpdateRoleReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = d.DMS.UpdateRole(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route DELETE /v1/dms/roles/{role_uid} dms DelRole
//
// Delete a role.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (a *DMSController) DelRole(c echo.Context) error {
	req := &aV1.DelRoleReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = a.DMS.DelRole(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route GET /v1/dms/roles dms ListRoles
//
// List roles.
//
//	responses:
//	  200: body:ListRoleReply
//	  default: body:GenericResp
func (d *DMSController) ListRoles(c echo.Context) error {
	req := new(aV1.ListRoleReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := d.DMS.ListRoles(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route POST /v1/dms/projects/{project_uid}/members dms AddMember
//
// Add member.
//
//	responses:
//	  200: body:AddMemberReply
//	  default: body:GenericResp
func (d *DMSController) AddMember(c echo.Context) error {
	req := new(aV1.AddMemberReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, err := d.DMS.AddMember(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/members dms ListMembers
//
// List member, for front page.
//
//	responses:
//	  200: body:ListMemberReply
//	  default: body:GenericResp
func (d *DMSController) ListMembers(c echo.Context) error {
	req := new(aV1.ListMemberReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := d.DMS.ListMembers(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/members/internal dms ListMembersForInternal
//
// List members, for internal backend service.
//
//	responses:
//	  200: body:ListMembersForInternalReply
//	  default: body:GenericResp
func (d *DMSController) ListMembersForInternal(c echo.Context) error {
	req := new(dmsV1.ListMembersForInternalReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := d.DMS.ListMembersForInternal(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route PUT /v1/dms/projects/{project_uid}/members/{member_uid} dms UpdateMember
//
// Update a member.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (d *DMSController) UpdateMember(c echo.Context) error {
	req := new(aV1.UpdateMemberReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = d.DMS.UpdateMember(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route DELETE /v1/dms/projects/{project_uid}/members/{member_uid} dms DelMember
//
// Delete a member.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (a *DMSController) DelMember(c echo.Context) error {
	req := &aV1.DelMemberReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = a.DMS.DelMember(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route GET /v1/dms/projects/{project_uid}/member_groups dms ListMemberGroups
//
// List member group, for front page.
//
//	responses:
//	  200: body:ListMemberGroupsReply
//	  default: body:GenericResp
func (d *DMSController) ListMemberGroups(c echo.Context) error {
	req := new(aV1.ListMemberGroupsReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := d.DMS.ListMemberGroups(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/member_groups/{member_group_uid} dms GetMemberGroup
//
// Get member group, for front page.
//
//	responses:
//	  200: body:GetMemberGroupReply
//	  default: body:GenericResp
func (d *DMSController) GetMemberGroup(c echo.Context) error {
	req := new(aV1.GetMemberGroupReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := d.DMS.GetMemberGroup(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route POST /v1/dms/projects/{project_uid}/member_groups dms AddMemberGroup
//
// Add member group.
//
//	responses:
//	  200: body:AddMemberGroupReply
//	  default: body:GenericResp
func (d *DMSController) AddMemberGroup(c echo.Context) error {
	req := new(aV1.AddMemberGroupReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, err := d.DMS.AddMemberGroup(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route PUT /v1/dms/projects/{project_uid}/member_groups/{member_group_uid} dms UpdateMemberGroup
//
// update member group, for front page.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (d *DMSController) UpdateMemberGroup(c echo.Context) error {
	req := new(aV1.UpdateMemberGroupReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = d.DMS.UpdateMemberGroup(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkResp(c)
}

// swagger:route DELETE  /v1/dms/projects/{project_uid}/member_groups/{member_group_uid} dms DeleteMemberGroup
//
// delete member group, for front page.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (d *DMSController) DeleteMemberGroup(c echo.Context) error {
	req := new(aV1.DeleteMemberGroupReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = d.DMS.DeleteMemberGroup(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkResp(c)
}

// swagger:route GET /v1/dms/op_permissions dms ListOpPermissions
//
// List op permission.
//
//	responses:
//	  200: body:ListOpPermissionReply
//	  default: body:GenericResp
func (d *DMSController) ListOpPermissions(c echo.Context) error {
	req := new(aV1.ListOpPermissionReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := d.DMS.ListOpPermissions(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects dms ListProjects
//
// List projects.
//
//	responses:
//	  200: body:ListProjectReply
//	  default: body:GenericResp
func (d *DMSController) ListProjects(c echo.Context) error {
	req := new(dmsV1.ListProjectReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, err := d.DMS.ListProjects(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route POST /v1/dms/projects dms AddProject
//
// Add project.
//
//	responses:
//	  200: body:AddProjectReply
//	  default: body:GenericResp
func (d *DMSController) AddProject(c echo.Context) error {
	req := new(aV1.AddProjectReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, err := d.DMS.AddProject(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route DELETE /v1/dms/projects/{project_uid} dms DelProject
//
// Delete a project
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (a *DMSController) DelProject(c echo.Context) error {
	req := &aV1.DelProjectReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	err = a.DMS.DeleteProject(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route PUT /v1/dms/projects/{project_uid} dms UpdateProject
//
// update a project.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (a *DMSController) UpdateProject(c echo.Context) error {
	req := &aV1.UpdateProjectReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	err = a.DMS.UpdateProjectDesc(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route PUT /v1/dms/projects/{project_uid}/archive dms ArchiveProject
//
// Archive a project.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (a *DMSController) ArchiveProject(c echo.Context) error {
	req := &aV1.ArchiveProjectReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = a.DMS.ArchivedProject(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route PUT /v1/dms/projects/{project_uid}/unarchive dms UnarchiveProject
//
// Unarchive a project.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (a *DMSController) UnarchiveProject(c echo.Context) error {
	req := &aV1.UnarchiveProjectReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = a.DMS.UnarchiveProject(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route POST /v1/dms/proxy dms RegisterDMSProxyTarget
//
// Register dms proxy target.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (d *DMSController) RegisterDMSProxyTarget(c echo.Context) error {
	req := new(dmsV1.RegisterDMSProxyTargetReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = d.DMS.RegisterDMSProxyTarget(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkResp(c)
}

// swagger:route POST /v1/dms/plugin dms RegisterDMSPlugin
//
// Register dms plugin.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (d *DMSController) RegisterDMSPlugin(c echo.Context) error {
	req := new(dmsV1.RegisterDMSPluginReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = d.DMS.RegisterDMSPlugin(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkResp(c)
}

// swagger:route GET /v1/dms/configurations/oauth2 dms GetOauth2Configuration
//
// Get Oauth2 configuration.
//
//	responses:
//	  200: body:GetOauth2ConfigurationResDataReply
//	  default: body:GenericResp
func (d *DMSController) GetOauth2Configuration(c echo.Context) error {
	reply, err := d.DMS.GetOauth2Configuration(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route PATCH /v1/dms/configurations/oauth2 dms UpdateOauth2Configuration
//
// Update Oauth2 configuration..
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (d *DMSController) UpdateOauth2Configuration(c echo.Context) error {
	req := new(aV1.Oauth2ConfigurationReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	err = d.DMS.UpdateOauth2Configuration(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkResp(c)
}

// swagger:route GET /v1/dms/oauth2/tips dms GetOauth2Tips
//
// Get Oauth2 Tips.
//
//	responses:
//	  200: body:GetOauth2TipsReply
//	  default: body:GenericResp
func (d *DMSController) GetOauth2Tips(c echo.Context) error {
	reply, err := d.DMS.GetOauth2ConfigurationTip(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/oauth2/link
//
// Oauth2 Link.
func (d *DMSController) Oauth2Link(c echo.Context) error {
	uri, err := d.DMS.Oauth2Link(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return c.Redirect(http.StatusFound, uri)
}

// Oauth2Callback is a hidden interface for third-party platform callbacks for oauth2 verification
func (d *DMSController) Oauth2Callback(c echo.Context) error {
	req := new(aV1.Oauth2CallbackReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	uri, err := d.DMS.Oauth2Callback(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return c.Redirect(http.StatusFound, uri)
}

// swagger:route POST /v1/dms/oauth2/user/bind dms BindOauth2User
//
// Bind Oauth2 User.
//
//	responses:
//	  200: body:BindOauth2UserReply
//	  default: body:GenericResp
func (d *DMSController) BindOauth2User(c echo.Context) error {
	req := new(aV1.BindOauth2UserReq)
	err := bindAndValidateReq(c, req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	reply, err := d.DMS.BindOauth2User(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}

	return NewOkRespWithReply(c, reply)
}

// swagger:route PATCH /v1/dms/configurations/ldap dms UpdateLDAPConfiguration
//
// Update ldap configuration.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (d *DMSController) UpdateLDAPConfiguration(c echo.Context) error {
	req := new(aV1.UpdateLDAPConfigurationReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	err = d.DMS.UpdateLDAPConfiguration(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkResp(c)
}

// swagger:route GET /v1/dms/configurations/ldap dms GetLDAPConfiguration
//
// Get ldap configuration.
//
//	responses:
//	  200: body:GetLDAPConfigurationResDataReply
//	  default: body:GenericResp
func (d *DMSController) GetLDAPConfiguration(c echo.Context) error {
	reply, err := d.DMS.GetLDAPConfiguration(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/configurations/smtp dms GetSMTPConfiguration
//
// get smtp configuration.
//
//	responses:
//	  200: body:GetSMTPConfigurationReply
//	  default: body:GenericResp
func (d *DMSController) GetSMTPConfiguration(c echo.Context) error {
	reply, err := d.DMS.GetSMTPConfiguration(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route PATCH /v1/dms/configurations/smtp dms UpdateSMTPConfiguration
//
// Get smtp configuration.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (d *DMSController) UpdateSMTPConfiguration(c echo.Context) error {
	req := new(aV1.UpdateSMTPConfigurationReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	err = d.DMS.UpdateSMTPConfiguration(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkResp(c)
}

// swagger:route POST /v1/dms/configurations/smtp/test dms TestSMTPConfiguration
//
// test smtp configuration.
//
//	responses:
//	  200: body:TestSMTPConfigurationReply
//	  default: body:GenericResp
func (d *DMSController) TestSMTPConfiguration(c echo.Context) error {
	req := new(aV1.TestSMTPConfigurationReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := d.DMS.TestSMTPConfiguration(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/configurations/wechat dms GetWeChatConfiguration
//
// get wechat configuration.
//
//	responses:
//	  200: body:GetWeChatConfigurationReply
//	  default: body:GenericResp
func (d *DMSController) GetWeChatConfiguration(c echo.Context) error {
	reply, err := d.DMS.GetWeChatConfiguration(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route PATCH /v1/dms/configurations/wechat dms UpdateWeChatConfiguration
//
// update wechat configuration.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (d *DMSController) UpdateWeChatConfiguration(c echo.Context) error {
	req := new(aV1.UpdateWeChatConfigurationReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	err = d.DMS.UpdateWeChatConfiguration(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkResp(c)
}

// swagger:route POST /v1/dms/configurations/wechat/test dms TestWeChatConfiguration
//
// test wechat configuration.
//
//	responses:
//	  200: body:TestWeChatConfigurationReply
//	  default: body:GenericResp
func (d *DMSController) TestWeChatConfiguration(c echo.Context) error {
	req := new(aV1.TestWeChatConfigurationReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := d.DMS.TestWeChatConfiguration(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/configurations/feishu dms GetFeishuConfiguration
//
// get feishu configuration.
//
//	responses:
//	  200: body:GetFeishuConfigurationReply
//	  default: body:GenericResp
func (d *DMSController) GetFeishuConfiguration(c echo.Context) error {
	reply, err := d.DMS.GetFeishuConfiguration(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route PATCH /v1/dms/configurations/feishu dms UpdateFeishuConfiguration
//
// update feishu configuration.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (d *DMSController) UpdateFeishuConfiguration(c echo.Context) error {
	req := new(aV1.UpdateFeishuConfigurationReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	err = d.DMS.UpdateFeishuConfiguration(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkResp(c)
}

// swagger:route POST /v1/dms/configurations/feishu/test dms TestFeishuConfiguration
//
// test feishu configuration.
//
//	responses:
//	  200: body:TestFeishuConfigurationReply
//	  default: body:GenericResp
func (d *DMSController) TestFeishuConfig(c echo.Context) error {
	req := new(aV1.TestFeishuConfigurationReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := d.DMS.TestFeishuConfiguration(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/configurations/webhook dms GetWebHookConfiguration
//
// get webhook configuration.
//
//	responses:
//	  200: body:GetWebHookConfigurationReply
//	  default: body:GenericResp
func (d *DMSController) GetWebHookConfiguration(c echo.Context) error {
	reply, err := d.DMS.GetWebHookConfiguration(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route PATCH /v1/dms/configurations/webhook dms UpdateWebHookConfiguration
//
// update webhook configuration.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (d *DMSController) UpdateWebHookConfiguration(c echo.Context) error {
	req := new(aV1.UpdateWebHookConfigurationReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	err = d.DMS.UpdateWebHookConfiguration(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkResp(c)
}

// swagger:route POST /v1/dms/configurations/webhook/test dms TestWebHookConfiguration
//
// test webhook configuration.
//
//	responses:
//	  200: body:TestWebHookConfigurationReply
//	  default: body:GenericResp
func (d *DMSController) TestWebHookConfiguration(c echo.Context) error {

	reply, err := d.DMS.TestWebHookConfiguration(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route POST /v1/dms/notifications dms Notification
//
// notify message.
//
//	responses:
//	  200: body:NotificationReply
//	  default: body:GenericResp
func (d *DMSController) Notify(c echo.Context) error {
	req := new(dmsV1.NotificationReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	err = d.DMS.NotifyMessage(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkResp(c)
}

// swagger:route POST /v1/dms/webhooks dms WebHookSendMessage
//
// webhook send message.
//
//	responses:
//	  200: body:WebHookSendMessageReply
//	  default: body:GenericResp
func (d *DMSController) WebHookSendMessage(c echo.Context) error {
	req := new(dmsV1.WebHookSendMessageReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	err = d.DMS.WebHookSendMessage(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkResp(c)
}

// swagger:route GET /v1/dms/company_notice dms GetCompanyNotice
//
// get company notice info
//
//	responses:
//	  200: body:GetCompanyNoticeReply
//	  default: body:GenericResp
func (d *DMSController) GetCompanyNotice(c echo.Context) error {
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	reply, err := d.DMS.GetCompanyNotice(c.Request().Context(), currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route PATCH /v1/dms/company_notice dms UpdateCompanyNotice
//
// update company notice info
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (d *DMSController) UpdateCompanyNotice(c echo.Context) error {
	req := new(aV1.UpdateCompanyNoticeReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	err = d.DMS.UpdateCompanyNotice(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkResp(c)
}
