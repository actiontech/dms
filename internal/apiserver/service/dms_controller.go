package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/actiontech/dms/api"
	aV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/apiserver/conf"
	apiError "github.com/actiontech/dms/internal/apiserver/pkg/error"
	"github.com/actiontech/dms/internal/dms/pkg/constant"
	"github.com/actiontech/dms/internal/dms/service"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	dmsV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	"github.com/actiontech/dms/pkg/dms-common/api/jwt"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"

	"github.com/labstack/echo/v4"
)

type DMSController struct {
	DMS                *service.DMSService
	CloudbeaverService *service.CloudbeaverService

	log *utilLog.Helper
	shutdownCallback func() error
}

func NewDMSController(logger utilLog.Logger, opts *conf.DMSOptions, cbService *service.CloudbeaverService) (*DMSController, error) {
	dmsService, err := service.NewAndInitDMSService(logger, opts)
	if nil != err {
		return nil, fmt.Errorf("failed to init dms service: %v", err)
	}
	return &DMSController{
		// log:   log.NewHelper(log.With(logger, "module", "controller/DMS")),
		DMS:                dmsService,
		CloudbeaverService: cbService,
		log: utilLog.NewHelper(logger, utilLog.WithMessageKey("controller")),
		shutdownCallback: func() error {
			if err := dmsService.Shutdown(); err != nil {
				return err
			}
			return nil
		},
	}, nil
}

func (ctl *DMSController) Shutdown() error {
	if nil != ctl.shutdownCallback {
		return ctl.shutdownCallback()
	}
	return nil
}


// swagger:operation POST /v1/dms/projects/{project_uid}/environment_tags Project CreateEnvironmentTag
//
// Create a new environment tag.
//
// ---
// parameters:
//   - name: project_uid
//     description: project uid
//     in: path
//     required: true
//     type: string
//   - name: environment_name
//     description: the name of environment tag to be created
//     in: body
//     required: true
//     type: string
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) CreateEnvironmentTag(c echo.Context) error {
	req := new(aV1.CreateEnvironmentTagReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = ctl.DMS.CreateEnvironmentTag(c.Request().Context(), req.ProjectUID, currentUserUid, req.Name)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkResp(c)
}

// swagger:operation PUT /v1/dms/projects/{project_uid}/environment_tags/{environment_tag_uid} Project UpdateEnvironmentTag
//
// Update an existing environment tag.
//
// ---
// parameters:
//   - name: project_uid
//     description: project uid
//     in: path
//     required: true
//     type: string
//   - name: environment_tag_uid
//     description: environment tag id
//     in: path
//     required: true
//     type: string
//   - name: environment_name
//     description: the name of environment tag to be updated
//     required: true
//     in: body
//     type: string
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateEnvironmentTag(c echo.Context) error {
	req := new(aV1.UpdateEnvironmentTagReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = ctl.DMS.UpdateEnvironmentTag(c.Request().Context(), req.ProjectUID, currentUserUid, req.EnvironmentTagUID, req.Name)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route DELETE /v1/dms/projects/{project_uid}/environment_tags/{environment_tag_uid} Project DeleteEnvironmentTag
//
// Delete an existing environment tag.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (ctl *DMSController) DeleteEnvironmentTag(c echo.Context) error {
	req := new(aV1.DeleteEnvironmentTagReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = ctl.DMS.DeleteEnvironmentTag(c.Request().Context(), req.ProjectUID, currentUserUid, req.EnvironmentTagUID)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route GET /v1/dms/projects/{project_uid}/environment_tags Project ListEnvironmentTags
//
// List environment tags.
//
//	responses:
//	  200: body:ListEnvironmentTagsReply
//	  default: body:GenericResp
func (ctl *DMSController) ListEnvironmentTags(c echo.Context) error{
	req := new(aV1.ListEnvironmentTagReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := ctl.DMS.ListEnvironmentTags(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/projects/{project_uid}/db_services DBService AddDBService
//
// Add DB Service.
//
// ---
// deprecated: true
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: db_service
//     description: Add new db service
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/AddDBServiceReq"
// responses:
//   '200':
//     description: AddDBServiceReply
//     schema:
//       "$ref": "#/definitions/AddDBServiceReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) AddDBService(c echo.Context) error {
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

	reply, err := ctl.DMS.AddDBService(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/db_services DBService ListDBServices
//
// List db service.
//
//	responses:
//	  200: body:ListDBServiceReply
//	  default: body:GenericResp
// deprecated: true
func (ctl *DMSController) ListDBServices(c echo.Context) error {
	return NewOkRespWithReply(c, nil)
}

// swagger:route GET /v1/dms/projects/{project_uid}/db_services/tips DBService ListDBServiceTips
//
// List db service tip.
//
//	responses:
//	  200: body:ListDBServiceTipsReply
//	  default: body:GenericResp
func (ctl *DMSController) ListDBServiceTips(c echo.Context) error {
	req := new(aV1.ListDBServiceTipsReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	reply, err := ctl.DMS.ListDBServiceTips(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/db_services/driver_options DBService ListDBServiceDriverOption
//
// List db service driver option.
//
//	responses:
//	  200: body:ListDBServiceDriverOptionReply
//	  default: body:GenericResp
func (ctl *DMSController) ListDBServiceDriverOption(c echo.Context) error {
	reply, err := ctl.DMS.ListDBServiceDriverOption(c.Request().Context())
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/db_services DBService ListGlobalDBServices
//
// list global DBServices
//
//	responses:
//	  200: body:ListGlobalDBServicesReply
//	  default: body:GenericResp
// deprecated: true
func (ctl *DMSController) ListGlobalDBServices(c echo.Context) error {
	return NewOkRespWithReply(c, nil)
}

// swagger:route GET /v1/dms/db_services/tips DBService ListGlobalDBServicesTips
//
// list global DBServices tips
//
//	responses:
//	  200: body:ListGlobalDBServicesTipsReply
//	  default: body:GenericResp
func (ctl *DMSController) ListGlobalDBServicesTips(c echo.Context) error {
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	reply, err := ctl.DMS.ListGlobalDBServicesTips(c.Request().Context(), currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route DELETE /v1/dms/projects/{project_uid}/db_services/{db_service_uid} DBService DelDBService
//
// Delete a DB Service.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (ctl *DMSController) DelDBService(c echo.Context) error {
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
	err = ctl.DMS.DelDBService(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:operation PUT /v1/dms/projects/{project_uid}/db_services/{db_service_uid} DBService UpdateDBService
//
// update a DB Service.
//
// ---
// deprecated: true
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: db_service_uid
//     description: db_service_uid id
//     in: path
//     required: true
//     type: string
//   - name: db_service
//     description: Update a DB service
//     in: body
//     schema:
//       "$ref": "#/definitions/UpdateDBServiceReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateDBService(c echo.Context) error {
	
	return NewOkResp(c)
}

// swagger:operation POST /v1/dms/projects/{project_uid}/db_services/connection DBService CheckDBServiceIsConnectable
//
// check if the db_service is connectable.
//
// ---
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: db_service
//     in: body
//     description: check db_service is connectable
//     schema:
//       "$ref": "#/definitions/CheckDBServiceIsConnectableReq"
// responses:
//   '200':
//     description: CheckDBServiceIsConnectableReply
//     schema:
//       "$ref": "#/definitions/CheckDBServiceIsConnectableReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) CheckDBServiceIsConnectable(c echo.Context) error {
	var req aV1.CheckDBServiceIsConnectableReq
	err := bindAndValidateReq(c, &req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.CheckDBServiceIsConnectable(c.Request().Context(), &req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/projects/{project_uid}/db_services/{db_service_uid}/connection DBService CheckDBServiceIsConnectableById
//
// check if the db_service is connectable.
//
// ---
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: db_service_uid
//     description: db service uid
//     in: path
//     required: true
//     type: string
// responses:
//   '200':
//     description: CheckDBServiceIsConnectableReply
//     schema:
//       "$ref": "#/definitions/CheckDBServiceIsConnectableReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) CheckDBServiceIsConnectableById(c echo.Context) error {
	var req aV1.CheckDBServiceIsConnectableByIdReq
	err := bindAndValidateReq(c, &req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.CheckDBServiceIsConnectableById(c.Request().Context(), &req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/projects/{project_uid}/db_services/connections DBService CheckProjectDBServicesConnections
//
// check if the project db_services is connectable.
//
// ---
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: db_services
//     description: check db_services is connectable
//     in: body
//     schema:
//       "$ref": "#/definitions/CheckDBServicesIsConnectableReq"
// responses:
//   '200':
//     description: CheckDBServicesIsConnectableReply
//     schema:
//       "$ref": "#/definitions/CheckDBServicesIsConnectableReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) CheckProjectDBServicesConnections(c echo.Context) error {
	var req aV1.CheckDBServicesIsConnectableReq
	err := bindAndValidateReq(c, &req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, err := ctl.DMS.CheckDBServiceIsConnectableByIds(c.Request().Context(), req.ProjectUid,currentUserUid,req.DBServices)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkRespWithReply(c, reply)
}


// swagger:route GET /v1/dms/basic_info BasicInfo GetBasicInfo
//
// get basic info.
//
//	responses:
//	  200: body:GetBasicInfoReply
//	  default: body:GenericResp
func (ctl *DMSController) GetBasicInfo(c echo.Context) error {
	reply, err := ctl.DMS.GetBasicInfo(c.Request().Context())
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/personalization/logo BasicInfo GetStaticLogo
//
// get logo
//
//	Produces:
//	- application/octet-stream
//
//	responses:
//	  200: GetStaticLogoReply
//	  default: body:GenericResp
func (ctl *DMSController) GetStaticLogo(c echo.Context) error {
	reply, contentType, err := ctl.DMS.GetStaticLogo(c.Request().Context())
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return c.Blob(http.StatusOK, contentType, reply.File)
}

// swagger:operation POST /v1/dms/personalization BasicInfo Personalization
//
// personalize [title, logo].
//
// ---
// parameters:
//   - name: title
//     description: title
//     in: formData
//     required: false
//     type: string
//   - name: file
//     description: file upload
//     in: formData
//     required: false
//     type: file
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) Personalization(c echo.Context) error {
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

	err = ctl.DMS.Personalization(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkResp(c)
}

// swagger:operation POST /v1/dms/users/verify_user_login User VerifyUserLogin
//
// Verify user login.
//
// ---
// parameters:
//   - name: session
//     in: body
//     required: true
//     description: Add a new session
//     schema:
//       "$ref": "#/definitions/AddSessionReq"
// responses:
//   '200':
//     description: VerifyUserLoginReply
//     schema:
//       "$ref": "#/definitions/VerifyUserLoginReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) VerifyUserLogin(c echo.Context) error {
	req := new(aV1.AddSessionReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := ctl.DMS.VerifyUserLogin(c.Request().Context(), &aV1.VerifyUserLoginReq{
		UserName: req.Session.UserName,
		Password: req.Session.Password,
	})
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/sessions Session AddSession
//
// Add a session.
//
// ---
// parameters:
//   - name: session
//     in: body
//     required: true
//     description: Add a new session
//     schema:
//       "$ref": "#/definitions/AddSessionReq"
// responses:
//   '200':
//     description: AddSessionReply
//     schema:
//       "$ref": "#/definitions/AddSessionReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) AddSession(c echo.Context) error {
	req := new(aV1.AddSessionReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := ctl.DMS.VerifyUserLogin(c.Request().Context(), &aV1.VerifyUserLoginReq{
		UserName: req.Session.UserName,
		Password: req.Session.Password,
		VerifyCode: req.Session.VerifyCode,
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
	refreshToken, err := jwt.GenRefreshToken(jwt.WithUserId(reply.Data.UserUid))
	if nil != err {
		return NewErrResp(c, err, apiError.APIServerErr)
	}

	err = ctl.DMS.AfterUserLogin(c.Request().Context(), &aV1.AfterUserLoginReq{
		UserUid: reply.Data.UserUid,
	})
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	c.SetCookie(&http.Cookie{
		Name:    constant.DMSToken,
		Value:   token,
		Path:    "/",
		Expires: time.Now().Add(jwt.DefaultDmsTokenExpHours * time.Hour),
	})
	c.SetCookie(&http.Cookie{
		Name:    constant.DMSRefreshToken,
		Value:   refreshToken,
		Path:    "/",
		HttpOnly: true, // 增加安全性
		SameSite:  http.SameSiteStrictMode, // cookie只会在同站请求中发送。
		Expires: time.Now().Add(jwt.DefaultDmsRefreshTokenExpHours * time.Hour),
	})

	return NewOkRespWithReply(c, &aV1.AddSessionReply{
		Data: struct {
			Token string `json:"token"`
			Message string `json:"message"`
		}{
			Token: token,
		},
	})
}

// swagger:route DELETE /v1/dms/sessions Session DelSession
//
// del a session.
//
//	responses:
//	  200: body:DelSessionReply
//	  default: body:GenericResp
func (ctl *DMSController) DelSession(c echo.Context) error {
	var redirectUri string

	refreshToken, err := c.Cookie(constant.DMSRefreshToken)
	if err != nil {
		ctl.log.Warnf("DelSession get refresh token cookie failed: %v, will not logout third-party platform session", err)
	} else {
		_,sub, sid, _, err := jwt.ParseRefreshToken(refreshToken.Value)
		if err != nil {
			ctl.log.Errorf("DelSession parse refresh token failed: %v, will not logout third-party platform session", err)
		} else {
			// 包含第三方会话信息，同步注销第三方平台会话
			redirectUri, err = ctl.DMS.Oauth2ConfigurationUsecase.Logout(c.Request().Context(), sub, sid)
			if err != nil {
				return NewErrResp(c, err, apiError.DMSServiceErr)
			}
		}
	}

	cookie, err := c.Cookie(constant.DMSToken)
	if err != nil {
		ctl.log.Warnf("DelSession get dms token cookie failed: %v", err)
	} else {
		// cookie 未过期
		cookie.MaxAge = -1 // MaxAge<0 means delete cookie now
		cookie.Path = "/"
		c.SetCookie(cookie)
		ctl.CloudbeaverService.Logout(cookie.Value)
	}

	reply := &aV1.DelSessionReply{Data: struct {
		Location string `json:"location"`
	}{Location: redirectUri}}
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false) // 避免将location中的 & 编码为 \u0026
	if err = enc.Encode(reply); err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}

	return c.JSONBlob(http.StatusOK, buf.Bytes())
}

// swagger:operation POST /v1/dms/sessions/refresh Session RefreshSession
//
// refresh a session.
//
// ---
// responses:
//   '200':
//     description: RefreshSession reply
//     schema:
//       "$ref": "#/definitions/AddSessionReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) RefreshSession(c echo.Context) error {
	refreshToken, err := c.Cookie(constant.DMSRefreshToken)
	if err != nil {
		return c.String(http.StatusUnauthorized, "refresh token not found")
	}

	uid, sub, sid, expired, err := jwt.ParseRefreshToken(refreshToken.Value)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	if expired {
		// 刷新token过期时，且包含第三方平台会话信息，注销第三方平台会话
		err = ctl.DMS.Oauth2ConfigurationUsecase.BackendLogout(c.Request().Context(), sub, sid)
		if err != nil {
			ctl.log.Errorf("expired refresh token, call BackendLogout err: %v", err)
			return NewErrResp(c, err, apiError.APIServerErr)
		}
		return c.String(http.StatusUnauthorized, "refresh token is expired")
	}

	// 签发的token包含第三方平台信息，需要同步刷新第三方平台token
	if sub != "" || sid != "" {
		claims, err := ctl.DMS.RefreshOauth2Token(c.Request().Context(), uid, sub, sid)
		if err != nil {
			return c.String(http.StatusUnauthorized, err.Error())
		}

		newDmsToken, dmsCookieExp, err := claims.DmsToken()
		if err != nil {
			return NewErrResp(c, err, apiError.APIServerErr)
		}
		newRefreshToken, dmsRefreshCookieExp, err := claims.DmsRefreshToken()
		if err != nil {
			return NewErrResp(c, err, apiError.APIServerErr)
		}

		c.SetCookie(&http.Cookie{
			Name:    constant.DMSToken,
			Value:   newDmsToken,
			Path:    "/",
			Expires: time.Now().Add(dmsCookieExp),
		})
		c.SetCookie(&http.Cookie{
			Name:    constant.DMSRefreshToken,
			Value:   newRefreshToken,
			Path:    "/",
			HttpOnly: true, // 增加安全性
			SameSite: http.SameSiteStrictMode, // cookie只会在同站请求中发送。
			Expires: time.Now().Add(dmsRefreshCookieExp),
		})

		return NewOkRespWithReply(c, &aV1.AddSessionReply{
			Data: struct {
				Token string `json:"token"`
				Message string `json:"message"`
			}{
				Token: newDmsToken,
			},
		})
	}

	// Create token with claims
	token, err := jwt.GenJwtToken(jwt.WithUserId(uid))
	if nil != err {
		return NewErrResp(c, err, apiError.APIServerErr)
	}

	c.SetCookie(&http.Cookie{
		Name:    constant.DMSToken,
		Value:   token,
		Path:    "/",
		Expires: time.Now().Add(jwt.DefaultDmsTokenExpHours * time.Hour),
	})

	return NewOkRespWithReply(c, &aV1.AddSessionReply{
		Data: struct {
			Token string `json:"token"`
			Message string `json:"message"`
		}{
			Token: token,
		},
	})
}

// swagger:route GET /v1/dms/sessions/user Session GetUserBySession
//
// Get current user.
//
//	responses:
//	  200: body:GetUserBySessionReply
//	  default: body:GenericResp
func (ctl *DMSController) GetUserBySession(c echo.Context) error {
	uid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := ctl.DMS.GetCurrentUser(c.Request().Context(), &aV1.GetUserBySessionReq{UserUid: uid})
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/users User AddUser
//
// Add user.
//
// ---
// parameters:
//   - name: user
//     in: body
//     required: true
//     description: Add new user
//     schema:
//       "$ref": "#/definitions/AddUserReq"
// responses:
//   '200':
//     description: AddUserReply
//     schema:
//       "$ref": "#/definitions/AddUserReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) AddUser(c echo.Context) error {
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

	reply, err := ctl.DMS.AddUser(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation PUT /v1/dms/users/{user_uid} User UpdateUser
//
// Update a user.
//
// ---
// parameters:
//   - name: user_uid
//     description: User uid
//     in: path
//     required: true
//     type: string
//   - name: user
//     description: Update a user
//     in: body
//     schema:
//       "$ref": "#/definitions/UpdateUserReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateUser(c echo.Context) error {
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

	err = ctl.DMS.UpdateUser(c.Request().Context(), req, currentUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:operation PUT /v1/dms/users User UpdateCurrentUser
//
// Update current user.
//
// ---
// parameters:
//   - name: current_user
//     description: Update current user
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/UpdateCurrentUserReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateCurrentUser(c echo.Context) error {
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

	err = ctl.DMS.UpdateCurrentUser(c.Request().Context(), req, currentUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route DELETE /v1/dms/users/{user_uid} User DelUser
//
// Delete a user.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (ctl *DMSController) DelUser(c echo.Context) error {
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

	err = ctl.DMS.DelUser(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}


// swagger:route GET /v1/dms/users User ListUsers
//
// List users.
//
//	responses:
//	  200: body:ListUserReply
//	  default: body:GenericResp
func (ctl *DMSController) ListUsers(c echo.Context) error {
	req := new(dmsV1.ListUserReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.ListUsers(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/users/{user_uid}/op_permission User GetUserOpPermission
//
// Get user op permission info, This API is used by other component such as sqle&auth to check user permissions.
//
//	responses:
//	  200: body:GetUserOpPermissionReply
//	  default: body:GenericResp
func (ctl *DMSController) GetUserOpPermission(c echo.Context) error {
	req := new(dmsV1.GetUserOpPermissionReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := ctl.DMS.GetUserOpPermission(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/users/{user_uid} User GetUser
//
// Get user info, This API is used by other component such as sqle&auth to get user info.
//
//	responses:
//	  200: body:GetUserReply
//	  default: body:GenericResp
func (ctl *DMSController) GetUser(c echo.Context) error {
	req := new(dmsV1.GetUserReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := ctl.DMS.GetUser(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/users/gen_token User GenAccessToken
//
// Gen user access token.
//
// ---
// parameters:
//   - name: expiration_days
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/GenAccessToken"
// responses:
//   '200':
//     description: GenAccessTokenReply
//     schema:
//       "$ref": "#/definitions/GenAccessTokenReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) GenAccessToken(c echo.Context) error {
	req := new(dmsV1.GenAccessToken)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, err := ctl.DMS.GenAccessToken(c.Request().Context(), currentUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/user_groups UserGroup AddUserGroup
//
// Add user group.
//
// ---
// parameters:
//   - name: user_group
//     description: Add new user group
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/AddUserGroupReq"
// responses:
//   '200':
//     description: AddUserGroupReply
//     schema:
//       "$ref": "#/definitions/AddUserGroupReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) AddUserGroup(c echo.Context) error {
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

	reply, err := ctl.DMS.AddUserGroup(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation PUT /v1/dms/user_groups/{user_group_uid} UserGroup UpdateUserGroup
//
// Update a user group.
//
// ---
// parameters:
//   - name: user_group_uid
//     description: UserGroup uid
//     in: path
//     required: true
//     type: string
//   - name: user_group
//     description: Update a user group
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/UpdateUserGroupReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateUserGroup(c echo.Context) error {
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

	err = ctl.DMS.UpdateUserGroup(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route DELETE /v1/dms/user_groups/{user_group_uid} UserGroup DelUserGroup
//
// Delete a user group.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (ctl *DMSController) DelUserGroup(c echo.Context) error {
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

	err = ctl.DMS.DelUserGroup(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route GET /v1/dms/user_groups UserGroup ListUserGroups
//
// List user groups.
//
//	responses:
//	  200: body:ListUserGroupReply
//	  default: body:GenericResp
func (ctl *DMSController) ListUserGroups(c echo.Context) error {
	req := new(aV1.ListUserGroupReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.ListUserGroups(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/roles Role AddRole
//
// Add role.
//
// ---
// parameters:
//   - name: role
//     description: Add new role
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/AddRoleReq"
// responses:
//   '200':
//     description: AddRoleReply
//     schema:
//       "$ref": "#/definitions/AddRoleReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) AddRole(c echo.Context) error {
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

	reply, err := ctl.DMS.AddRole(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation PUT /v1/dms/roles/{role_uid} Role UpdateRole
//
// Update a role.
//
// ---
// parameters:
//   - name: role_uid
//     description: Role uid
//     in: path
//     required: true
//     type: string
//   - name: role
//     description: Update a role
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/UpdateRoleReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateRole(c echo.Context) error {
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

	err = ctl.DMS.UpdateRole(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route DELETE /v1/dms/roles/{role_uid} Role DelRole
//
// Delete a role.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (ctl *DMSController) DelRole(c echo.Context) error {
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

	err = ctl.DMS.DelRole(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route GET /v1/dms/roles Role ListRoles
//
// List roles.
//
//	responses:
//	  200: body:ListRoleReply
//	  default: body:GenericResp
func (ctl *DMSController) ListRoles(c echo.Context) error {
	req := new(aV1.ListRoleReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.ListRoles(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/projects/{project_uid}/members Member AddMember
//
// Add member.
//
// ---
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: member
//     description: Add new member
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/AddMemberReq"
// responses:
//   '200':
//     description: AddMemberReply
//     schema:
//       "$ref": "#/definitions/AddMemberReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) AddMember(c echo.Context) error {
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

	reply, err := ctl.DMS.AddMember(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/members/tips Member ListMemberTips
//
// List member tips.
//
//	responses:
//	  200: body:ListMemberTipsReply
//	  default: body:GenericResp
func (ctl *DMSController) ListMemberTips(c echo.Context) error {
	req := new(aV1.ListMemberTipsReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.ListMemberTips(c.Request().Context(), req.ProjectUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/members Member ListMembers
//
// List member, for front page.
//
//	responses:
//	  200: body:ListMemberReply
//	  default: body:GenericResp
func (ctl *DMSController) ListMembers(c echo.Context) error {
	req := new(aV1.ListMemberReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	reply, err := ctl.DMS.ListMembers(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/members/internal Member ListMembersForInternal
//
// List members, for internal backend service.
//
//	responses:
//	  200: body:ListMembersForInternalReply
//	  default: body:GenericResp
func (ctl *DMSController) ListMembersForInternal(c echo.Context) error {
	req := new(dmsV1.ListMembersForInternalReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.ListMembersForInternal(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation PUT /v1/dms/projects/{project_uid}/members/{member_uid} Member UpdateMember
//
// Update a member.
//
// ---
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: member_uid
//     description: Member uid
//     in: path
//     required: true
//     type: string
//   - name: member
//     description: Update a member
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/UpdateMemberReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateMember(c echo.Context) error {
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

	err = ctl.DMS.UpdateMember(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route DELETE /v1/dms/projects/{project_uid}/members/{member_uid} Member DelMember
//
// Delete a member.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (ctl *DMSController) DelMember(c echo.Context) error {
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

	err = ctl.DMS.DelMember(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route GET /v1/dms/projects/{project_uid}/member_groups MemberGroup ListMemberGroups
//
// List member group, for front page.
//
//	responses:
//	  200: body:ListMemberGroupsReply
//	  default: body:GenericResp
func (ctl *DMSController) ListMemberGroups(c echo.Context) error {
	req := new(aV1.ListMemberGroupsReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	reply, err := ctl.DMS.ListMemberGroups(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/member_groups/tips MemberGroup ListMemberGroupTips
//
// List member group tips.
//
//	responses:
//	  200: body:ListMemberGroupTipsReply
//	  default: body:GenericResp
func (ctl *DMSController) ListMemberGroupTips(c echo.Context) error {
	req := new(aV1.ListMemberGroupTipsReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.ListMemberGroupTips(c.Request().Context(), req.ProjectUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/member_groups/{member_group_uid} MemberGroup GetMemberGroup
//
// Get member group, for front page.
//
//	responses:
//	  200: body:GetMemberGroupReply
//	  default: body:GenericResp
func (ctl *DMSController) GetMemberGroup(c echo.Context) error {
	req := new(aV1.GetMemberGroupReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := ctl.DMS.GetMemberGroup(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/projects/{project_uid}/member_groups MemberGroup AddMemberGroup
//
// Add member group.
//
// ---
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: member_group
//     description: Add new member group
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/AddMemberGroupReq"
// responses:
//   '200':
//     description: AddMemberGroupReply
//     schema:
//       "$ref": "#/definitions/AddMemberGroupReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) AddMemberGroup(c echo.Context) error {
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

	reply, err := ctl.DMS.AddMemberGroup(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation PUT /v1/dms/projects/{project_uid}/member_groups/{member_group_uid} MemberGroup UpdateMemberGroup
//
// update member group, for front page.
//
// ---
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: member_group_uid
//     description: Member group id
//     in: path
//     required: true
//     type: string
//   - name: member_group
//     description: Update a member group
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/UpdateMemberGroupReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateMemberGroup(c echo.Context) error {
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

	err = ctl.DMS.UpdateMemberGroup(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkResp(c)
}

// swagger:route DELETE  /v1/dms/projects/{project_uid}/member_groups/{member_group_uid} MemberGroup DeleteMemberGroup
//
// delete member group, for front page.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (ctl *DMSController) DeleteMemberGroup(c echo.Context) error {
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

	err = ctl.DMS.DeleteMemberGroup(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkResp(c)
}

// swagger:route GET /v1/dms/op_permissions OpPermission ListOpPermissions
//
// List op permission.
//
//	responses:
//	  200: body:ListOpPermissionReply
//	  default: body:GenericResp
func (ctl *DMSController) ListOpPermissions(c echo.Context) error {
	req := new(aV1.ListOpPermissionReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.ListOpPermissions(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects Project ListProjects
//
// List projects.
//
//	responses:
//	  200: body:ListProjectReply
//	  default: body:GenericResp
// deprecated: true
func (ctl *DMSController) ListProjects(c echo.Context) error {
	return nil
}

// swagger:operation POST /v1/dms/projects/business_tags Project CreateBusinessTag
//
// Create a new business tag.
//
// ---
// parameters:
//   - name: business_tag
//     description: business tag to be created
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/CreateBusinessTagReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) CreateBusinessTag(c echo.Context) error {
	req := new(aV1.CreateBusinessTagReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = ctl.DMS.CreateBusinessTag(c.Request().Context(), currentUserUid, req.BusinessTag)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkResp(c)
}

// swagger:operation PUT /v1/dms/projects/business_tags/{business_tag_uid} Project UpdateBusinessTag
//
// Update an existing business tag.
//
// ---
// parameters:
//   - name: business_tag_uid
//     description: business tag id
//     in: path
//     required: true
//     type: string
//   - name: business_tag
//     description: the business tag to be updated
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/UpdateBusinessTagReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateBusinessTag(c echo.Context) error {
	req := new(aV1.UpdateBusinessTagReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = ctl.DMS.UpdateBusinessTag(c.Request().Context(), currentUserUid, req.BusinessTagUID, req.BusinessTag)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route DELETE /v1/dms/projects/business_tags/{business_tag_uid} Project DeleteBusinessTag
//
// Delete an existing business tag.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (ctl *DMSController) DeleteBusinessTag(c echo.Context) error {
	req := new(aV1.DeleteBusinessTagReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = ctl.DMS.DeleteBusinessTag(c.Request().Context(), currentUserUid, req.BusinessTagUID)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route GET /v1/dms/projects/business_tags Project ListBusinessTags
//
// List business tags.
//
//	responses:
//	  200: body:ListBusinessTagsReply
//	  default: body:GenericResp
func (ctl *DMSController) ListBusinessTags(c echo.Context) error{
	req := new(aV1.ListBusinessTagReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := ctl.DMS.ListBusinessTags(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/projects Project AddProject
//
// Add project.
//
// ---
// deprecated: true
// parameters:
//   - name: project
//     description: Add new Project
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/AddProjectReq"
// responses:
//   '200':
//     description: AddProjectReply
//     schema:
//       "$ref": "#/definitions/AddProjectReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) AddProject(c echo.Context) error {
	return nil
}

// swagger:route DELETE /v1/dms/projects/{project_uid} Project DelProject
//
// Delete a project
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (ctl *DMSController) DelProject(c echo.Context) error {
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
	err = ctl.DMS.DeleteProject(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:operation PUT /v1/dms/projects/{project_uid} Project UpdateProject
//
// update a project.
//
// ---
// deprecated: true
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: project
//     description: Update a project
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/UpdateProjectReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateProject(c echo.Context) error {
	return nil
}

// swagger:route PUT /v1/dms/projects/{project_uid}/archive Project ArchiveProject
//
// Archive a project.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (ctl *DMSController) ArchiveProject(c echo.Context) error {
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

	err = ctl.DMS.ArchivedProject(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route PUT /v1/dms/projects/{project_uid}/unarchive Project UnarchiveProject
//
// Unarchive a project.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (ctl *DMSController) UnarchiveProject(c echo.Context) error {
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

	err = ctl.DMS.UnarchiveProject(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:operation POST /v1/dms/projects/import Project ImportProjects
//
// Import projects.
//
// ---
// deprecated: true
// parameters:
//   - name: projects
//     description: import projects
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/ImportProjectsReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) ImportProjects(c echo.Context) error {
	return nil 
}

// swagger:route POST /v1/dms/projects/preview_import Project PreviewImportProjects
//
// Preview import projects.
//
//	Consumes:
//	- multipart/form-data
//
//	responses:
//	  200: PreviewImportProjectsReply
//	  default: body:GenericResp
// deprecated: true
func (ctl *DMSController) PreviewImportProjects(c echo.Context) error {
	return nil
}

// swagger:route GET /v1/dms/projects/import_template Project GetImportProjectsTemplate
//
// Get import projects template.
//
//	responses:
//	  200: GetImportProjectsTemplateReply
//	  default: body:GenericResp
func (ctl *DMSController) GetImportProjectsTemplate(c echo.Context) error {
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	content, err := ctl.DMS.GetImportProjectsTemplate(c.Request().Context(), currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	c.Response().Header().Set(echo.HeaderContentDisposition,
		mime.FormatMediaType("attachment", map[string]string{"filename": "导入项目模版.csv"}))

	return c.Blob(http.StatusOK, "text/csv", content)
}

// swagger:route GET /v1/dms/projects/export Project ExportProjects
//
// Export projects file.
//
//	responses:
//	  200: ExportProjectsReply
//	  default: body:GenericResp
func (ctl *DMSController) ExportProjects(c echo.Context) error {
	req := new(aV1.ExportProjectsReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	content, err := ctl.DMS.ExportProjects(c.Request().Context(), currentUserUid, req)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	fileName := fmt.Sprintf("项目列表_%s.csv", time.Now().Format("20060102150405"))
	c.Response().Header().Set(echo.HeaderContentDisposition,
		mime.FormatMediaType("attachment", map[string]string{"filename": fileName}))

	return c.Blob(http.StatusOK, "text/csv", content)
}

// swagger:route GET /v1/dms/projects/tips Project GetProjectTips
//
// Get project tips.
//
//	responses:
//	  200: body:GetProjectTipsReply
//	  default: body:GenericResp
// deprecated: true
func (ctl *DMSController) GetProjectTips(c echo.Context) error {
	return nil
}

// swagger:route GET /v1/dms/projects/import_db_services_template Project GetImportDBServicesTemplate
//
// Get import DBServices template.
//
//	responses:
//	  200: GetImportDBServicesTemplateReply
//	  default: body:GenericResp
func (ctl *DMSController) GetImportDBServicesTemplate(c echo.Context) error {
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	importDBServicesTemplate, err := ctl.DMS.GetImportDBServicesTemplate(c.Request().Context(), currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	c.Response().Header().Set(echo.HeaderContentDisposition,
		mime.FormatMediaType("attachment", map[string]string{"filename": "import_db_services_template.csv"}))

	return c.Blob(http.StatusOK, "text/csv", importDBServicesTemplate)
}

// swagger:route POST /v1/dms/projects/{project_uid}/db_services/import_check DBService ImportDBServicesOfOneProjectCheck
//
// Import DBServices.
//
//	Consumes:
//	- multipart/form-data
//
//	Produces:
//	- application/json
//	- text/csv
//
//	responses:
//	  200: ImportDBServicesCheckCsvReply
//	  default: body:ImportDBServicesCheckReply
// deprecated: true
func (ctl *DMSController) ImportDBServicesOfOneProjectCheck(c echo.Context) error {
	return NewOkRespWithReply(c, nil)
}

// swagger:operation POST /v1/dms/projects/{project_uid}/db_services/import DBService ImportDBServicesOfOneProject
//
// Import DBServices.
//
// ---
// deprecated: true
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: db_services
//     description: new db services
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/ImportDBServicesOfOneProjectReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) ImportDBServicesOfOneProject(c echo.Context) error {
	return NewOkResp(c)
}

// swagger:route POST /v1/dms/projects/import_db_services_check Project ImportDBServicesOfProjectsCheck
//
// Import DBServices.
//
//		Consumes:
//		- multipart/form-data
//
//		Produces:
//		- application/json
//		- text/csv
//
//	responses:
//	  200: ImportDBServicesCheckCsvReply
//	  default: body:ImportDBServicesCheckReply
// deprecated: true
func (ctl *DMSController) ImportDBServicesOfProjectsCheck(c echo.Context) error {
	return NewOkRespWithReply(c, nil)
}

// swagger:operation POST /v1/dms/projects/import_db_services Project ImportDBServicesOfProjects
//
// Import DBServices.
//
// ---
// parameters:
//   - name: db_services
//     description: new db services
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/ImportDBServicesOfProjectsReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
// deprecated: true
func (ctl *DMSController) ImportDBServicesOfProjects(c echo.Context) error {
	return NewOkResp(c)
}

// todo 该接口已废弃
// swagger:operation POST /v1/dms/projects/db_services_connection Project DBServicesConnection
//
// DBServices Connection.
//
// ---
// parameters:
//   - name: db_services
//     description: check db_service is connectable
//     in: body
//     schema:
//       "$ref": "#/definitions/DBServiceConnectionReq"
// responses:
//   '200':
//     description: DBServicesConnectionReply
//     schema:
//       "$ref": "#/definitions/DBServicesConnectionReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) DBServicesConnection(c echo.Context) error {
	req := new(aV1.DBServiceConnectionReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, err := ctl.DMS.DBServicesConnection(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/projects/db_services_connections Project CheckGlobalDBServicesConnections
//
// check if the global db_services is connectable.
//
// ---
// parameters:
//   - name: db_services
//     description: check db_services is connectable
//     in: body
//     schema:
//       "$ref": "#/definitions/DBServicesConnectionReq"
// responses:
//   '200':
//     description: DBServicesConnectionReqReply
//     schema:
//       "$ref": "#/definitions/DBServicesConnectionReqReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) CheckGlobalDBServicesConnections(c echo.Context) error {
	var req aV1.DBServicesConnectionReq
	err := bindAndValidateReq(c, &req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, err := ctl.DMS.CheckDBServiceIsConnectableByIds(c.Request().Context(),"", currentUserUid,req.DBServices)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkRespWithReply(c, reply)
}


// swagger:operation POST /v1/dms/projects/check_db_services_privileges Project CheckDBServicesPrivileges
//
// check if the db_services hava enough privileges.
//
// ---
// parameters:
//   - name: db_services
//     in: body
//     description: check db_services have enough privileges
//     schema:
//       "$ref": "#/definitions/CheckDBServicesPrivilegesReq"
// responses:
//   '200':
//     description: CheckDBServicesPrivilegesReply
//     schema:
//       "$ref": "#/definitions/CheckDBServicesPrivilegesReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) CheckDBServicesPrivileges(c echo.Context) error {
	var req aV1.CheckDBServicesPrivilegesReq
	err := bindAndValidateReq(c, &req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.CheckDBServiceHasEnoughPrivileges(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}


// swagger:operation POST /v1/dms/proxys DMSProxy RegisterDMSProxyTarget
//
// Register dms proxy target.
//
// ---
// parameters:
//   - name: dms_proxy_target
//     description: register dms proxy
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/RegisterDMSProxyTargetReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) RegisterDMSProxyTarget(c echo.Context) error {
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

	err = ctl.DMS.RegisterDMSProxyTarget(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkResp(c)
}

// swagger:operation POST /v1/dms/plugins DMSPlugin RegisterDMSPlugin
//
// Register dms plugin.
//
// ---
// parameters:
//   - name: plugin
//     description: Register dms plugin
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/RegisterDMSPluginReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) RegisterDMSPlugin(c echo.Context) error {
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

	err = ctl.DMS.RegisterDMSPlugin(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkResp(c)
}

// swagger:route GET /v1/dms/configurations/login/tips Configuration GetLoginTips
//
// get login configuration.
//
//	responses:
//	  200: body:GetLoginTipsReply
//	  default: body:GenericResp
func (ctl *DMSController) GetLoginTips(c echo.Context) error {
	reply, err := ctl.DMS.GetLoginTips(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation PATCH /v1/dms/configurations/login Configuration UpdateLoginConfiguration
//
// Update login configuration.
//
// ---
// parameters:
//   - name: login
//     description: update login configuration
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/UpdateLoginConfigurationReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateLoginConfiguration(c echo.Context) error {
	req := new(aV1.UpdateLoginConfigurationReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = ctl.DMS.UpdateLoginConfiguration(c.Request().Context(), currentUserUid, req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkResp(c)
}

// swagger:route GET /v1/dms/configurations/oauth2 Configuration GetOauth2Configuration
//
// Get Oauth2 configuration.
//
//	responses:
//	  200: body:GetOauth2ConfigurationResDataReply
//	  default: body:GenericResp
func (ctl *DMSController) GetOauth2Configuration(c echo.Context) error {
	reply, err := ctl.DMS.GetOauth2Configuration(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation PATCH /v1/dms/configurations/oauth2 Configuration UpdateOauth2Configuration
//
// Update Oauth2 configuration..
//
// ---
// parameters:
//   - name: oauth2
//     description: update oauth2 configuration
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/Oauth2ConfigurationReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateOauth2Configuration(c echo.Context) error {
	req := new(aV1.Oauth2ConfigurationReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	err = ctl.DMS.UpdateOauth2Configuration(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkResp(c)
}

// swagger:route GET /v1/dms/oauth2/tips OAuth2 GetOauth2Tips
//
// Get Oauth2 Tips.
//
//	responses:
//	  200: body:GetOauth2TipsReply
//	  default: body:GenericResp
func (ctl *DMSController) GetOauth2Tips(c echo.Context) error {
	reply, err := ctl.DMS.GetOauth2ConfigurationTip(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/oauth2/link OAuth2 Oauth2LinkOrCallback
//
// Oauth2 Link or Callback.
func (ctl *DMSController) Oauth2LinkOrCallback(c echo.Context) error {
	// 如果是OAuth2 callback请求，QueryParam中会有code这个参数
	if c.QueryParam("code") == "" {
		return ctl.oauth2Link(c)
	}
	return ctl.oauth2Callback(c)
}

func (ctl *DMSController) oauth2Link(c echo.Context) error {
	uri, err := ctl.DMS.Oauth2Link(c.Request().Context(), c.QueryString())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return c.Redirect(http.StatusFound, uri)
}

// oauth2Callback is a hidden interface for third-party platform callbacks for oauth2 verification
func (ctl *DMSController) oauth2Callback(c echo.Context) error {
	req := new(aV1.Oauth2CallbackReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	callbackData, claims, err := ctl.DMS.Oauth2Callback(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}

	// 1. callbackData.Error 有错误时，前端会回到登录页并展示错误信息
	// 2. callbackData.UserExist 为false时，前端会进入手动绑定页面，绑定时调用绑定接口签发tokens
	// 3. 没错误且用户存在时，签发tokens登录成功
	if  callbackData.Error == "" && callbackData.UserExist {
		dmsToken, dmsCookieExp, err := claims.DmsToken()
		if err != nil {
			return NewErrResp(c, err, apiError.APIServerErr)
		}
		refreshToken, dmsRefreshCookieExp, err := claims.DmsRefreshToken()
		if err != nil {
			return NewErrResp(c, err, apiError.APIServerErr)
		}
		callbackData.State = req.State
		callbackData.DMSToken = dmsToken
		c.SetCookie(&http.Cookie{
			Name:    constant.DMSToken,
			Value:   dmsToken,
			Path:    "/",
			Expires: time.Now().Add(dmsCookieExp),
		})
		c.SetCookie(&http.Cookie{
			Name:    constant.DMSRefreshToken,
			Value:   refreshToken,
			Path:    "/",
			HttpOnly: true, // 增加安全性
			SameSite:  http.SameSiteStrictMode, // cookie只会在同站请求中发送。
			Expires: time.Now().Add(dmsRefreshCookieExp),
		})
	}


	return c.Redirect(http.StatusFound, callbackData.Generate())
}

// swagger:operation POST /v1/dms/oauth2/user/bind OAuth2 BindOauth2User
//
// Bind Oauth2 User.
//
// ---
// parameters:
//   - name: BindOauth2UserReq
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/BindOauth2UserReq"
// responses:
//   '200':
//     description: BindOauth2UserReply
//     schema:
//       "$ref": "#/definitions/BindOauth2UserReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) BindOauth2User(c echo.Context) error {
	req := new(aV1.BindOauth2UserReq)
	err := bindAndValidateReq(c, req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	claims, err := ctl.DMS.BindOauth2User(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}

	dmsToken, dmsCookieExp, err := claims.DmsToken()
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	refreshToken, dmsRefreshCookieExp, err := claims.DmsRefreshToken()
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}

	c.SetCookie(&http.Cookie{
		Name:    constant.DMSToken,
		Value:   dmsToken,
		Path:    "/",
		Expires: time.Now().Add(dmsCookieExp),
	})
	c.SetCookie(&http.Cookie{
		Name:    constant.DMSRefreshToken,
		Value:   refreshToken,
		Path:    "/",
		HttpOnly: true, // 增加安全性
		SameSite:  http.SameSiteStrictMode, // cookie只会在同站请求中发送。
		Expires: time.Now().Add(dmsRefreshCookieExp),
	})
		
	return NewOkRespWithReply(c, &aV1.BindOauth2UserReply{
		Data:aV1.BindOauth2UserResData{Token: dmsToken},
	})
}

// BackChannelLogout is a hidden interface for third-party platform callbacks for logout event
// https://openid.net/specs/openid-connect-backchannel-1_0.html#BCRequest
func (ctl *DMSController) BackChannelLogout(c echo.Context) error {
	// no-store 指令告诉浏览器和任何中间缓存（例如代理服务器）不要存储响应的任何副本。
	// 这意味着每次请求该资源时，都必须从服务器重新获取
	c.Response().Header().Set(echo.HeaderCacheControl, "no-store")
	if err := c.Request().ParseForm(); err != nil {
		return c.String(http.StatusBadRequest, "Invalid form data")
	}

	logoutToken := c.Request().Form.Get("logout_token")
	if logoutToken == "" {
		return c.String(http.StatusBadRequest, "Missing logout_token")
	}

	// todo Verifier logoutToken by provider

	err := ctl.DMS.BackChannelLogout(c.Request().Context(), logoutToken)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

// swagger:operation PATCH /v1/dms/configurations/ldap Configuration UpdateLDAPConfiguration
//
// Update ldap configuration.
//
// ---
// parameters:
//   - name: ldap
//     description: update ldap configuration
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/UpdateLDAPConfigurationReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateLDAPConfiguration(c echo.Context) error {
	req := new(aV1.UpdateLDAPConfigurationReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	err = ctl.DMS.UpdateLDAPConfiguration(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkResp(c)
}

// swagger:route GET /v1/dms/configurations/ldap Configuration GetLDAPConfiguration
//
// Get ldap configuration.
//
//	responses:
//	  200: body:GetLDAPConfigurationResDataReply
//	  default: body:GenericResp
func (ctl *DMSController) GetLDAPConfiguration(c echo.Context) error {
	reply, err := ctl.DMS.GetLDAPConfiguration(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/configurations/smtp Configuration GetSMTPConfiguration
//
// get smtp configuration.
//
//	responses:
//	  200: body:GetSMTPConfigurationReply
//	  default: body:GenericResp
func (ctl *DMSController) GetSMTPConfiguration(c echo.Context) error {
	reply, err := ctl.DMS.GetSMTPConfiguration(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation PATCH /v1/dms/configurations/smtp Configuration UpdateSMTPConfiguration
//
// Update smtp configuration.
//
// ---
// parameters:
//   - name: smtp_configuration
//     description: update smtp configuration
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/UpdateSMTPConfigurationReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateSMTPConfiguration(c echo.Context) error {
	req := new(aV1.UpdateSMTPConfigurationReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	err = ctl.DMS.UpdateSMTPConfiguration(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkResp(c)
}

// swagger:operation POST /v1/dms/configurations/smtp/test Configuration TestSMTPConfiguration
//
// test smtp configuration.
//
// ---
// parameters:
//   - name: test_smtp_configuration
//     description: test smtp configuration
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/TestSMTPConfigurationReq"
// responses:
//   '200':
//     description: TestSMTPConfigurationReply
//     schema:
//       "$ref": "#/definitions/TestSMTPConfigurationReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) TestSMTPConfiguration(c echo.Context) error {
	req := new(aV1.TestSMTPConfigurationReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := ctl.DMS.TestSMTPConfiguration(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/configurations/wechat Configuration GetWeChatConfiguration
//
// get wechat configuration.
//
//	responses:
//	  200: body:GetWeChatConfigurationReply
//	  default: body:GenericResp
func (ctl *DMSController) GetWeChatConfiguration(c echo.Context) error {
	reply, err := ctl.DMS.GetWeChatConfiguration(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation PATCH /v1/dms/configurations/wechat Configuration UpdateWeChatConfiguration
//
// update wechat configuration.
//
// ---
// parameters:
//   - name: update_wechat_configuration
//     description: update wechat configuration
//     in: body
//     schema:
//       "$ref": "#/definitions/UpdateWeChatConfigurationReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateWeChatConfiguration(c echo.Context) error {
	req := new(aV1.UpdateWeChatConfigurationReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	err = ctl.DMS.UpdateWeChatConfiguration(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkResp(c)
}

// swagger:operation POST /v1/dms/configurations/wechat/test Configuration TestWeChatConfiguration
//
// test wechat configuration.
//
// ---
// parameters:
//   - name: test_wechat_configuration
//     description: test wechat configuration
//     in: body
//     schema:
//       "$ref": "#/definitions/TestWeChatConfigurationReq"
// responses:
//   '200':
//     description: TestWeChatConfigurationReply
//     schema:
//       "$ref": "#/definitions/TestWeChatConfigurationReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) TestWeChatConfiguration(c echo.Context) error {
	req := new(aV1.TestWeChatConfigurationReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := ctl.DMS.TestWeChatConfiguration(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/configurations/feishu Configuration GetFeishuConfiguration
//
// get feishu configuration.
//
//	responses:
//	  200: body:GetFeishuConfigurationReply
//	  default: body:GenericResp
func (ctl *DMSController) GetFeishuConfiguration(c echo.Context) error {
	reply, err := ctl.DMS.GetFeishuConfiguration(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation PATCH /v1/dms/configurations/feishu Configuration UpdateFeishuConfiguration
//
// update feishu configuration.
//
// ---
// parameters:
//   - name: update_feishu_configuration
//     description: update feishu configuration
//     in: body
//     schema:
//       "$ref": "#/definitions/UpdateFeishuConfigurationReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateFeishuConfiguration(c echo.Context) error {
	req := new(aV1.UpdateFeishuConfigurationReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	err = ctl.DMS.UpdateFeishuConfiguration(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkResp(c)
}

// swagger:operation POST /v1/dms/configurations/feishu/test Configuration TestFeishuConfiguration
//
// test feishu configuration.
//
// ---
// parameters:
//   - name: test_feishu_configuration
//     description: test feishu configuration
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/TestFeishuConfigurationReq"
// responses:
//   '200':
//     description: TestFeishuConfigurationReply
//     schema:
//       "$ref": "#/definitions/TestFeishuConfigurationReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) TestFeishuConfig(c echo.Context) error {
	req := new(aV1.TestFeishuConfigurationReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := ctl.DMS.TestFeishuConfiguration(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/configurations/webhook Configuration GetWebHookConfiguration
//
// get webhook configuration.
//
//	responses:
//	  200: body:GetWebHookConfigurationReply
//	  default: body:GenericResp
func (ctl *DMSController) GetWebHookConfiguration(c echo.Context) error {
	reply, err := ctl.DMS.GetWebHookConfiguration(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation PATCH /v1/dms/configurations/webhook Configuration UpdateWebHookConfiguration
//
// update webhook configuration.
//
// ---
// parameters:
//   - name: webhook_config
//     description: webhook configuration
//     in: body
//     schema:
//       "$ref": "#/definitions/UpdateWebHookConfigurationReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateWebHookConfiguration(c echo.Context) error {
	req := new(aV1.UpdateWebHookConfigurationReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	err = ctl.DMS.UpdateWebHookConfiguration(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkResp(c)
}

// swagger:route POST /v1/dms/configurations/webhook/test Configuration TestWebHookConfiguration
//
// test webhook configuration.
//
//	responses:
//	  200: body:TestWebHookConfigurationReply
//	  default: body:GenericResp
func (ctl *DMSController) TestWebHookConfiguration(c echo.Context) error {

	reply, err := ctl.DMS.TestWebHookConfiguration(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation PATCH /v1/dms/configurations/sms Configuration UpdateSmsConfiguration
//
// update sms configuration.
//
// ---
// parameters:
//   - name: update_sms_configuration
//     description: update sms configuration
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/UpdateSmsConfigurationReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateSmsConfiguration(context echo.Context) error {
	req := new(aV1.UpdateSmsConfigurationReq)
	err := bindAndValidateReq(context, req)
	if nil != err {
		return NewErrResp(context, err, apiError.BadRequestErr)
	}
	err = ctl.DMS.UpdateSmsConfiguration(context.Request().Context(), req)
	if err != nil {
		return NewErrResp(context, err, apiError.APIServerErr)
	}
	return NewOkResp(context)
}

// swagger:operation POST /v1/dms/configurations/sms/test Configuration TestSmsConfiguration
//
// test smtp configuration.
//
// ---
// parameters:
//   - name: test_sms_configuration
//     description: test sms configuration
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/TestSmsConfigurationReq"
// responses:
//   '200':
//     description: TestSmsConfigurationReply
//     schema:
//       "$ref": "#/definitions/TestSmsConfigurationReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) TestSmsConfiguration(context echo.Context) error {
	req := new(aV1.TestSmsConfigurationReq)
	err := bindAndValidateReq(context, req)
	if nil != err {
		return NewErrResp(context, err, apiError.BadRequestErr)
	}
	reply, err := ctl.DMS.TestSmsConfiguration(context.Request().Context(), req)
	if err != nil {
		return NewErrResp(context, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(context, reply)
}


// swagger:route GET /v1/dms/configurations/sms Configuration GetSmsConfiguration
//
// get sms configuration.
//
//	responses:
//	  200: body:GetSmsConfigurationReply
//	  default: body:GenericResp
func (ctl *DMSController) GetSmsConfiguration(c echo.Context) error {
	reply, err := ctl.DMS.GetSmsConfiguration(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/configurations/sms/send_code SMS SendSmsCode
//
// send sms code.
//
// ---
// parameters:
//   - name: username
//     description: user name
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/SendSmsCodeReq"
// responses:
//   '200':
//     description: SendSmsCodeReply
//     schema:
//       "$ref": "#/definitions/SendSmsCodeReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) SendSmsCode(context echo.Context) error {
	req := new(aV1.SendSmsCodeReq)
	err := bindAndValidateReq(context, req)
	reply, err := ctl.DMS.SendSmsCode(context.Request().Context(), req.Username)
	if err != nil {
		return NewErrResp(context, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(context, reply)
}

// swagger:operation POST /v1/dms/configurations/sms/verify_code SMS VerifySmsCode
//
// verify sms code.
//
// ---
// parameters:
//   - name: code
//     description: verify sms code
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/VerifySmsCodeReq"
//   - name: username
//     description: user name
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/VerifySmsCodeReq"
// responses:
//   '200':
//     description: VerifySmsCodeReply
//     schema:
//       "$ref": "#/definitions/VerifySmsCodeReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) VerifySmsCode(context echo.Context) error {
	req := new(aV1.VerifySmsCodeReq)
	err := bindAndValidateReq(context, req)
	if nil != err {
		return NewErrResp(context, err, apiError.BadRequestErr)
	}
	reply :=ctl.DMS.VerifySmsCode(req)
	return NewOkRespWithReply(context, reply)
}

// swagger:route POST /v1/dms/notifications Notification Notification
//
// notify message.
//
//	responses:
//	  200: body:NotificationReply
//	  default: body:GenericResp
func (ctl *DMSController) Notify(c echo.Context) error {
	req := new(dmsV1.NotificationReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	err = ctl.DMS.NotifyMessage(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkResp(c)
}

// swagger:operation POST /v1/dms/webhooks Webhook WebHookSendMessage
//
// webhook send message.
//
// ---
// parameters:
//   - name: webhook_message
//     description: webhooks
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/WebHookSendMessageReq"
// responses:
//   '200':
//     description: WebHookSendMessageReply
//     schema:
//       "$ref": "#/definitions/WebHookSendMessageReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) WebHookSendMessage(c echo.Context) error {
	req := new(dmsV1.WebHookSendMessageReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	err = ctl.DMS.WebHookSendMessage(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkResp(c)
}

// swagger:route GET /v1/dms/company_notice CompanyNotice GetCompanyNotice
//
// get company notice info
//
//	responses:
//	  200: body:GetCompanyNoticeReply
//	  default: body:GenericResp
func (ctl *DMSController) GetCompanyNotice(c echo.Context) error {
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	reply, err := ctl.DMS.GetCompanyNotice(c.Request().Context(), currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation PATCH /v1/dms/company_notice CompanyNotice UpdateCompanyNotice
//
// update company notice info
//
// ---
// parameters:
//   - name: company_notice
//     description: Update a companynotice
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/UpdateCompanyNoticeReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateCompanyNotice(c echo.Context) error {
	req := new(aV1.UpdateCompanyNoticeReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	err = ctl.DMS.UpdateCompanyNotice(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkResp(c)
}

// swagger:route GET /v1/dms/configurations/license Configuration GetLicense
//
// get license.
//
//	responses:
//	  200: body:GetLicenseReply
//	  default: body:GenericResp
func (ctl *DMSController) GetLicense(c echo.Context) error {
	reply, err := ctl.DMS.GetLicense(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}

const (
	HardwareInfoFileName   = "collected.infos"
	LicenseFileParamKey    = "license_file"
	ProjectsFileParamKey   = "projects_file"
	DBServicesFileParamKey = "db_services_file"
)

// swagger:route GET /v1/dms/configurations/license/info Configuration GetLicenseInfo
//
// get generate license info.
//
//	responses:
//	  200: GetLicenseInfoReply
//	  default: body:GenericResp
func (ctl *DMSController) GetLicenseInfo(c echo.Context) error {
	data, err := ctl.DMS.GetLicenseInfo(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	c.Response().Header().Set(echo.HeaderContentDisposition,
		mime.FormatMediaType("attachment", map[string]string{"filename": HardwareInfoFileName}))

	return c.Blob(http.StatusOK, echo.MIMEOctetStream, []byte(data))
}

// swagger:route GET /v1/dms/configurations/license/usage Configuration GetLicenseUsage
//
// get license usage.
//
//	responses:
//	  200: body:GetLicenseUsageReply
//	  default: body:GenericResp
func (ctl *DMSController) GetLicenseUsage(c echo.Context) error {
	usage, err := ctl.DMS.GetLicenseUsage(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}

	return NewOkRespWithReply(c, usage)
}

// swagger:route POST /v1/dms/configurations/license Configuration SetLicense
//
// import license.
//
//	 Consumes:
//	 - multipart/form-data
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (ctl *DMSController) SetLicense(c echo.Context) error {
	file, exist, err := ReadFileContent(c, LicenseFileParamKey)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	if !exist {
		return NewErrResp(c, fmt.Errorf("upload file is not exist"), apiError.APIServerErr)
	}
	err = ctl.DMS.SetLicense(c.Request().Context(), file)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkResp(c)
}

// swagger:route POST /v1/dms/configurations/license/check Configuration CheckLicense
//
// notify message.
//
//	 Consumes:
//	 - multipart/form-data
//
//	responses:
//	  200: body:CheckLicenseReply
//	  default: body:GenericResp
func (ctl *DMSController) CheckLicense(c echo.Context) error {
	file, exist, err := ReadFileContent(c, LicenseFileParamKey)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	if !exist {
		return NewErrResp(c, fmt.Errorf("upload file is not exist"), apiError.APIServerErr)
	}

	reply, err := ctl.DMS.CheckLicense(c.Request().Context(), file)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}

	return NewOkRespWithReply(c, reply)
}

// ReadFileContent read content from http body by name if file exist,
// the name is a http form data key, not file name.
func ReadFileContent(c echo.Context, name string) (content string, fileExist bool, err error) {
	file, err := c.FormFile(name)
	if err == http.ErrMissingFile {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	src, err := file.Open()
	if err != nil {
		return "", false, err
	}
	defer src.Close()
	data, err := io.ReadAll(src)
	if err != nil {
		return "", false, err
	}
	return string(data), true, nil
}

// swagger:operation POST /v1/dms/projects/{project_uid}/data_export_workflows DataExportWorkflows AddDataExportWorkflow
//
// Add data_export workflow.
//
// ---
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: data_export_workflow
//     description: add data export workflow
//     in: body
//     schema:
//       "$ref": "#/definitions/AddDataExportWorkflowReq"
// responses:
//   '200':
//     description: AddDataExportWorkflowReply
//     schema:
//       "$ref": "#/definitions/AddDataExportWorkflowReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) AddDataExportWorkflow(c echo.Context) error {
	req := new(aV1.AddDataExportWorkflowReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, err := ctl.DMS.AddDataExportWorkflow(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route POST /v1/dms/projects/{project_uid}/data_export_workflows/{data_export_workflow_uid}/approve DataExportWorkflows ApproveDataExportWorkflow
//
// Approve data_export workflow.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (ctl *DMSController) ApproveDataExportWorkflow(c echo.Context) error {
	req := &aV1.ApproveDataExportWorkflowReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	if err = ctl.DMS.ApproveDataExportWorkflow(c.Request().Context(), req, currentUserUid); err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkResp(c)
}

// swagger:operation POST /v1/dms/projects/{project_uid}/data_export_workflows/{data_export_workflow_uid}/reject DataExportWorkflows RejectDataExportWorkflow
//
// Reject data_export workflow.
//
// ---
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: data_export_workflow_uid
//     in: path
//     required: true
//     type: string
//   - name: payload
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/RejectDataExportWorkflowReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) RejectDataExportWorkflow(c echo.Context) error {
	req := &aV1.RejectDataExportWorkflowReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	if err = ctl.DMS.RejectDataExportWorkflow(c.Request().Context(), req, currentUserUid); err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkResp(c)
}

// swagger:route GET /v1/dms/projects/{project_uid}/data_export_workflows DataExportWorkflows ListDataExportWorkflows
//
// List data_export workflow.
//
//	responses:
//	  200: body:ListDataExportWorkflowsReply
//	  default: body:GenericResp
func (ctl *DMSController) ListDataExportWorkflows(c echo.Context) error {
	req := new(aV1.ListDataExportWorkflowsReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	reply, err := ctl.DMS.ListDataExportWorkflow(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/data_export_workflows DataExportWorkflows ListDataExportWorkflows
//
// List all data_export workflow.
//
//	responses:
//	  200: body:ListDataExportWorkflowsReply
//	  default: body:GenericResp
func (ctl *DMSController) ListAllDataExportWorkflows(c echo.Context) error {
	req := new(aV1.ListDataExportWorkflowsReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	reply, err := ctl.DMS.ListDataExportWorkflow(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/data_export_workflows/{data_export_workflow_uid} DataExportWorkflows GetDataExportWorkflow
//
// Get data_export workflow.
//
//	responses:
//	  200: body:GetDataExportWorkflowReply
//	  default: body:GenericResp
func (ctl *DMSController) GetDataExportWorkflow(c echo.Context) error {
	req := new(aV1.GetDataExportWorkflowReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	reply, err := ctl.DMS.GetDataExportWorkflow(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/projects/{project_uid}/data_export_workflows/cancel DataExportWorkflows CancelDataExportWorkflow
//
// Cancel data export workflows.
//
// ---
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: payload
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/CancelDataExportWorkflowReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) CancelDataExportWorkflow(c echo.Context) error {
	req := &aV1.CancelDataExportWorkflowReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	if err = ctl.DMS.CancelDataExportWorkflow(c.Request().Context(), req, currentUserUid); err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkResp(c)
}

// swagger:route POST /v1/dms/projects/{project_uid}/data_export_workflows/{data_export_workflow_uid}/export DataExportWorkflows ExportDataExportWorkflow
//
// exec data_export workflow.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (ctl *DMSController) ExportDataExportWorkflow(c echo.Context) error {
	req := &aV1.ExportDataExportWorkflowReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	err = ctl.DMS.ExportDataExportWorkflow(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkResp(c)
}

// swagger:operation POST /v1/dms/projects/{project_uid}/data_export_tasks DataExportTask AddDataExportTask
//
// Add data_export task.
//
// ---
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: data_export_tasks
//     description: add data export workflow
//     in: body
//     schema:
//       "$ref": "#/definitions/AddDataExportTaskReq"
// responses:
//   '200':
//     description: AddDataExportTaskReply
//     schema:
//       "$ref": "#/definitions/AddDataExportTaskReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) AddDataExportTask(c echo.Context) error {
	req := new(aV1.AddDataExportTaskReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, err := ctl.DMS.AddDataExportTask(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)

}

// swagger:route GET /v1/dms/projects/{project_uid}/data_export_tasks DataExportTask BatchGetDataExportTask
//
// Batch get data_export task.
//
//	responses:
//	  200: body:BatchGetDataExportTaskReply
//	  default: body:GenericResp
func (ctl *DMSController) BatchGetDataExportTask(c echo.Context) error {
	req := new(aV1.BatchGetDataExportTaskReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.BatchGetDataExportTask(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/data_export_tasks/{data_export_task_uid}/data_export_task_sqls DataExportTask ListDataExportTaskSQLs
//
// List data_export workflow.
//
//	responses:
//	  200: body:ListDataExportTaskSQLsReply
//	  default: body:GenericResp
func (ctl *DMSController) ListDataExportTaskSQLs(c echo.Context) error {
	req := new(aV1.ListDataExportTaskSQLsReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	reply, err := ctl.DMS.ListDataExportTaskSQLs(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/data_export_tasks/{data_export_task_uid}/data_export_task_sqls/download DataExportTask DownloadDataExportTaskSQLs
//
// dowload data_export sqls.
//
//	responses:
//	  200: DownloadDataExportTaskSQLsReply
//	  default: body:GenericResp
func (ctl *DMSController) DownloadDataExportTaskSQLs(c echo.Context) error {
	req := new(aV1.DownloadDataExportTaskSQLsReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	fileName, content, err := ctl.DMS.DownloadDataExportTaskSQLs(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	c.Response().Header().Set(echo.HeaderContentDisposition,
		mime.FormatMediaType("attachment", map[string]string{"filename": fileName}))

	return c.Blob(http.StatusOK, echo.MIMETextPlain, content)
}

// swagger:route GET /v1/dms/projects/{project_uid}/data_export_tasks/{data_export_task_uid}/download DataExportTask DownloadDataExportTask
//
// download task file.
//
//	responses:
//	  200: DownloadDataExportTaskReply
//	  default: body:GenericResp
func (ctl *DMSController) DownloadDataExportTask(c echo.Context) error {
	req := &aV1.DownloadDataExportTaskReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	isProxy, filePath, err := ctl.DMS.DownloadDataExportTask(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	if isProxy {
		return ctl.proxyDownloadDataExportTask(c, filePath)
	}

	fileName := filepath.Base(filePath)
	c.Response().Header().Set(echo.HeaderContentDisposition,
		mime.FormatMediaType("attachment", map[string]string{"filename": fileName}))

	return c.File(filePath)
}

func (ctl *DMSController) proxyDownloadDataExportTask(c echo.Context, reportHost string) (err error) {
	protocol := strings.ToLower(strings.Split(c.Request().Proto, "/")[0])

	// reference from echo framework proxy middleware
	target, _ := url.Parse(fmt.Sprintf("%s://%s", protocol, reportHost))
	reverseProxy := httputil.NewSingleHostReverseProxy(target)
	reverseProxy.ErrorHandler = func(resp http.ResponseWriter, req *http.Request, err error) {
		// If the client canceled the request (usually by closing the connection), we can report a
		// client error (4xx) instead of a server error (5xx) to correctly identify the situation.
		// The Go standard library (at of late 2020) wraps the exported, standard
		// context.Canceled error with unexported garbage value requiring a substring check, see
		// https://github.com/golang/go/blob/6965b01ea248cabb70c3749fd218b36089a21efb/src/net/net.go#L416-L430
		if err == context.Canceled || strings.Contains(err.Error(), "operation was canceled") {
			httpError := echo.NewHTTPError(middleware.StatusCodeContextCanceled, fmt.Sprintf("client closed connection: %v", err))
			httpError.Internal = err
			c.Set("_error", httpError)
		} else {
			httpError := echo.NewHTTPError(http.StatusBadGateway, fmt.Sprintf("remote %s unreachable, could not forward: %v", reportHost, err))
			httpError.Internal = err
			c.Set("_error", httpError)
		}
	}

	reverseProxy.ServeHTTP(c.Response(), c.Request())

	if e, ok := c.Get("_error").(error); ok {
		err = e
	}

	return
}

// swagger:route GET /v1/dms/masking/rules Masking ListMaskingRules
//
// List masking rules.
//
//	responses:
//	  200: body:ListMaskingRulesReply
//	  default: body:GenericResp
func (ctl *DMSController) ListMaskingRules(c echo.Context) error {
	req := &aV1.ListMaskingRulesReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.ListMaskingRules(c.Request().Context())
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/cb_operation_logs CBOperationLogs ListCBOperationLogs
//
// List cb operation logs.
//
//	responses:
//	  200: body:ListCBOperationLogsReply
//	  default: body:GenericResp
func (ctl *DMSController) ListCBOperationLogs(c echo.Context) error {
	req := &aV1.ListCBOperationLogsReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, err := ctl.DMS.ListCBOperationLogs(c.Request().Context(), req, currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}

	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/cb_operation_logs/export CBOperationLogs ExportCBOperationLogs
//
// Export cb operation logs.
//
//	responses:
//	  200: ExportCBOperationLogsReply
//	  default: body:GenericResp
func (ctl *DMSController) ExportCBOperationLogs(c echo.Context) error {
	req := &aV1.ExportCBOperationLogsReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	content, err := ctl.DMS.ExportCBOperationLogs(c.Request().Context(), req, currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}

	fileName := fmt.Sprintf("CBoperation_%s.csv", time.Now().Format("20060102150405.000"))
	c.Response().Header().Set(echo.HeaderContentDisposition,
		mime.FormatMediaType("attachment", map[string]string{"filename": fileName}))

	return c.Blob(http.StatusOK, "text/csv", content)

}

// swagger:route GET /v1/dms/projects/{project_uid}/cb_operation_logs/tips CBOperationLogs GetCBOperationLogTips
//
// Get cb operation log tips.
//
//	responses:
//	  200: GetCBOperationLogTipsReply
//	  default: body:GenericResp
func (ctl *DMSController) GetCBOperationLogTips(c echo.Context) error {
	req := &aV1.GetCBOperationLogTipsReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, err := ctl.DMS.GetCBOperationLogTips(c.Request().Context(), req, currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}

	return NewOkRespWithReply(c, reply)
}

func (ctl *DMSController) SwaggerHandler(c echo.Context) error {
	err := ctl.DMS.RegisterSwagger(c)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}

	optionList := []func(config *echoSwagger.Config){
		func(config *echoSwagger.Config) {
			// for clear the default URLs
			config.URLs = []string{}
		},
	}

	// 设置InstanceName,为了找到正确的swagger配置
	for swagType := range api.GetAllSwaggerDocs() {
		urlPath := swagType.GetUrlPath()
		optionList = append(optionList, echoSwagger.URL(urlPath))

		if strings.HasSuffix(c.Request().RequestURI, urlPath) {
			optionList = append(optionList, echoSwagger.InstanceName(urlPath))
		}
	}

	handler := echoSwagger.EchoWrapHandler(optionList...)
	return handler(c)
}

// swagger:operation GET /v1/dms/db_service_sync_tasks DBServiceSyncTask ListDBServiceSyncTasks
//
// List database synchronization tasks.
//
// ---
// responses:
//   '200':
//     description: ListDBServiceSyncTasksReply
//     schema:
//       "$ref": "#/definitions/ListDBServiceSyncTasksReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) ListDBServiceSyncTasks(c echo.Context) error {
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	reply, err := ctl.DMS.ListDBServiceSyncTask(c.Request().Context(), currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation GET /v1/dms/db_service_sync_tasks/{db_service_sync_task_uid} DBServiceSyncTask GetDBServiceSyncTask
//
// Get database synchronization task.
//
// ---
// parameters:
//   - name: db_service_sync_task_uid
//     description: db service sync task uid
//     in: path
//     required: true
//     type: string
// responses:
//   '200':
//     description: GetDBServiceSyncTaskReply
//     schema:
//       "$ref": "#/definitions/GetDBServiceSyncTaskReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) GetDBServiceSyncTask(c echo.Context) error {
	req := new(aV1.GetDBServiceSyncTaskReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	reply, err := ctl.DMS.GetDBServiceSyncTask(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/db_service_sync_tasks DBServiceSyncTask AddDBServiceSyncTask
//
// Add database synchronization task.
//
// ---
// parameters:
//   - name: db_service_sync_task
//     description: Add new db service sync task
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/AddDBServiceSyncTaskReq"
// responses:
//   '200':
//     description: AddDBServiceReply
//     schema:
//       "$ref": "#/definitions/AddDBServiceReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) AddDBServiceSyncTask(c echo.Context) error {
	req := new(aV1.AddDBServiceSyncTaskReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, err := ctl.DMS.AddDBServiceSyncTask(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation PUT /v1/dms/db_service_sync_tasks/{db_service_sync_task_uid} DBServiceSyncTask UpdateDBServiceSyncTask
//
// update database synchronization task.
//
// ---
// parameters:
//   - name: db_service_sync_task_uid
//     description: db service sync task uid
//     in: path
//     required: true
//     type: string
//   - name: db_service_sync_task
//     description: update db service sync task
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/UpdateDBServiceSyncTaskReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateDBServiceSyncTask(c echo.Context) error {
	req := &aV1.UpdateDBServiceSyncTaskReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	err = ctl.DMS.UpdateDBServiceSyncTask(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:operation DELETE /v1/dms/db_service_sync_tasks/{db_service_sync_task_uid} DBServiceSyncTask DeleteDBServiceSyncTask
//
// Delete database synchronization task.
//
// ---
// parameters:
//   - name: db_service_sync_task_uid
//     description: db service sync task uid
//     in: path
//     required: true
//     type: string
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) DeleteDBServiceSyncTask(c echo.Context) error {
	req := &aV1.DeleteDBServiceSyncTaskReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	err = ctl.DMS.DeleteDBServiceSyncTask(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:operation GET /v1/dms/db_service_sync_tasks/tips DBServiceSyncTask ListDBServiceSyncTaskTips
//
// List database synchronization task tips.
//
// ---
// responses:
//   '200':
//     description: ListDBServiceSyncTaskTipsReply
//     schema:
//       "$ref": "#/definitions/ListDBServiceSyncTaskTipsReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) ListDBServiceSyncTaskTips(c echo.Context) error {
	reply, err := ctl.DMS.ListDBServiceSyncTaskTips(c.Request().Context())
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/db_service_sync_tasks/{db_service_sync_task_uid}/sync DBServiceSyncTask SyncDBServices
//
// Sync db service.
//
// ---
// parameters:
//   - name: db_service_sync_task_uid
//     description: db service sync task uid
//     in: path
//     required: true
//     type: string
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) SyncDBServices(c echo.Context) error {
	req := &aV1.SyncDBServicesReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	err = ctl.DMS.SyncDBServices(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkResp(c)
}

// swagger:operation GET /v1/dms/resource_overview/statistics ResourceOverview GetResourceOverviewStatisticsV1
//
// Get resource overview statistics.
//
// ---
// responses:
//   '200':
//     description: resource overview statistics response body
//     schema:
//       "$ref": "#/definitions/ResourceOverviewStatisticsResV1"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (d *DMSController) GetResourceOverviewStatistics(c echo.Context) error {
	// 获取当前用户ID
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	// 直接获取并返回统计信息
	reply, err := d.DMS.GetResourceOverviewStatistics(c.Request().Context(), currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkRespWithReply(c, reply)
}

// swagger:operation GET /v1/dms/resource_overview/resource_type_distribution ResourceOverview GetResourceOverviewResourceTypeDistributionV1
//
// Get resource overview resource type distribution.
//
// ---
// responses:
//   '200':
//     description: resource overview resource type distribution response body
//     schema:
//       "$ref": "#/definitions/ResourceOverviewResourceTypeDistributionResV1"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (d *DMSController) GetResourceOverviewResourceTypeDistribution(c echo.Context) error {
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	// 直接获取并返回统计信息
	reply, err := d.DMS.GetResourceOverviewResourceTypeDistribution(c.Request().Context(), currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkRespWithReply(c, reply)
}

// swagger:operation GET /v1/dms/resource_overview/topology ResourceOverview GetResourceOverviewTopologyV1
//
// Get resource overview topology.
//
// ---
// responses:
//   '200':
//     description: resource overview topology response body
//     schema:
//       "$ref": "#/definitions/ResourceOverviewTopologyResV1"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (d *DMSController) GetResourceOverviewTopology(c echo.Context) error {
	req := &aV1.ResourceOverviewTopologyReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	// 获取资源概览拓扑图
	reply, err := d.DMS.GetResourceOverviewTopology(c.Request().Context(), currentUserUid, req)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkRespWithReply(c, reply)
}

// swagger:operation GET /v1/dms/resource_overview/resource_list ResourceOverview GetResourceOverviewResourceListV1
//
// Get resource overview resource list.
//
// ---
// responses:
//   '200':
//     description: resource overview resource list response body
//     schema:
//       "$ref": "#/definitions/ResourceOverviewResourceListResV1"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (d *DMSController) GetResourceOverviewResourceList(c echo.Context) error {
	req := &aV1.ResourceOverviewResourceListReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	// 获取资源概详情列表
	reply, err := d.DMS.GetResourceOverviewList(c.Request().Context(), currentUserUid, req)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/resource_overview/download ResourceOverview DownloadResourceOverviewList
//
// download resource overview list csv file.
//
//	responses:
//	  200: DownloadResourceOverviewListRes
//	  default: body:GenericResp
func (d *DMSController) DownloadResourceOverviewList(c echo.Context) error {
	req := &aV1.DownloadResourceOverviewListReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	content, err := d.DMS.DownloadResourceOverviewList(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	// TODO 后续需要增加国际化时，需要修改文件名以及CSV文件中的字段
	fileName := fmt.Sprintf("资源全景视图_%s.csv", time.Now().Format("20060102150405"))
	c.Response().Header().Set(echo.HeaderContentDisposition,
		mime.FormatMediaType("attachment", map[string]string{"filename": fileName}))

	return c.Blob(http.StatusOK, "text/csv", content)
}




// swagger:operation PUT /v1/dms/gateways/{gateway_id} Gateway UpdateGateway
//
// update gateways.
//
// ---
// parameters:
//   - name: gateway_id
//     description: gateway id
//     in: path
//     required: true
//     type: string
//   - name: update_gateway
//     description: update gateway
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/UpdateGatewayReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateGateway(c echo.Context) error {
	req := &aV1.UpdateGatewayReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	err = ctl.DMS.UpdateGateway(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

func (ctl *DMSController) SyncGateways(c echo.Context) error {
	req := &aV1.SyncGatewayReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	err = ctl.DMS.SyncGateways(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route GET /v1/dms/gateways Gateway ListGateways
//
// List gateways.
//
//	responses:
//	  200: body:ListGatewaysReply
//	  default: body:GenericResp
func (ctl *DMSController) ListGateways(c echo.Context) error {
	req := &aV1.ListGatewaysReq{}
	err := bindAndValidateReq(c, req)
	if nil!= err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := ctl.DMS.ListGateways(c.Request().Context(),req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/gateways/{gateway_id} Gateway GetGateway
//
// Get gateways.
//
//	responses:
//	  200: body:GetGatewayReply
//	  default: body:GenericResp
func (ctl *DMSController)GetGateway(c echo.Context) error {
	req := &aV1.GetGatewayReq{}
	err := bindAndValidateReq(c, req)
	if nil!= err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := ctl.DMS.GetGateway(c.Request().Context(), req)
	if nil!= err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/gateways/tips Gateway GetGatewayTips
//
// Get gateway tips.
//
//	responses:
//	  200: body:GetGatewayTipsReply
//	  default: body:GenericResp
func (ctl *DMSController)GetGatewayTips(c echo.Context) error {
	reply,err := ctl.DMS.GetGatewayTips(c.Request().Context())
	if nil!= err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route DELETE /v1/dms/gateways/{gateway_id} Gateway DeleteGateway
//
// Delete gateways.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (ctl *DMSController)DeleteGateway(c echo.Context) error {
	req := &aV1.DeleteGatewayReq{}
	err := bindAndValidateReq(c, req)
	if nil!= err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	err = ctl.DMS.DeleteGateway(c.Request().Context(), req)
	if nil!= err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:operation POST /v1/dms/gateways Gateway AddGateway
//
// Add gateways.
//
// ---
// parameters:
//   - name: add_gateway
//     description: add gateway
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/AddGatewayReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController)AddGateway(c echo.Context) error {
	req := &aV1.AddGatewayReq{}
	err := bindAndValidateReq(c, req)
	if nil!= err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	err = ctl.DMS.AddGateway(c.Request().Context(), req.AddGateway)
	if nil!= err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route GET /v1/dms/configurations/system_variables Configuration GetSystemVariables
//
// 获取系统变量配置
//
//	responses:
//	  200: body:GetSystemVariablesReply
//	  default: body:GenericResp
func (ctl *DMSController) GetSystemVariables(c echo.Context) error {
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	reply, err := ctl.DMS.GetSystemVariables(c.Request().Context(), currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation PATCH /v1/dms/configurations/system_variables Configuration UpdateSystemVariables
//
// 更新系统变量配置
//
// ---
// parameters:
//   - name: system_variables
//     description: 更新系统变量配置
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/UpdateSystemVariablesReqV1"
//
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateSystemVariables(c echo.Context) error {
	req := new(dmsV1.UpdateSystemVariablesReqV1)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = ctl.DMS.UpdateSystemVariables(c.Request().Context(), req, currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}
