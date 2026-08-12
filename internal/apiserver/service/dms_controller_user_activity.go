package service

import (
	aV1 "github.com/actiontech/dms/api/dms/service/v1"
	apiError "github.com/actiontech/dms/internal/apiserver/pkg/error"
	"github.com/actiontech/dms/pkg/dms-common/api/jwt"
	"github.com/labstack/echo/v4"
)

// swagger:operation GET /v1/dms/statistic/user_activity/summary UserActivity GetUserActivitySummary
//
// Get user activity summary KPI for a date.
//
// ---
// parameters:
//   - name: stat_date
//     in: query
//     required: true
//     type: string
// responses:
//   '200':
//     description: GetUserActivitySummaryReply
//     schema:
//       "$ref": "#/definitions/GetUserActivitySummaryReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) GetUserActivitySummary(c echo.Context) error {
	req := new(aV1.GetUserActivitySummaryReq)
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	reply, err := ctl.DMS.GetUserActivitySummary(c.Request().Context(), req, currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation GET /v1/dms/statistic/user_activity/daily_trend UserActivity ListUserActivityDailyTrend
//
// List user activity daily trend.
//
// ---
// parameters:
//   - name: filter_date_from
//     in: query
//     required: true
//     type: string
//   - name: filter_date_to
//     in: query
//     required: true
//     type: string
// responses:
//   '200':
//     description: ListUserActivityDailyTrendReply
//     schema:
//       "$ref": "#/definitions/ListUserActivityDailyTrendReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) ListUserActivityDailyTrend(c echo.Context) error {
	req := new(aV1.ListUserActivityDailyTrendReq)
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	reply, err := ctl.DMS.ListUserActivityDailyTrend(c.Request().Context(), req, currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation GET /v1/dms/statistic/user_activity/module_distribution UserActivity ListUserActivityModuleDistribution
//
// List user activity module distribution for a date.
//
// ---
// parameters:
//   - name: stat_date
//     in: query
//     required: true
//     type: string
// responses:
//   '200':
//     description: ListUserActivityModuleDistributionReply
//     schema:
//       "$ref": "#/definitions/ListUserActivityModuleDistributionReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) ListUserActivityModuleDistribution(c echo.Context) error {
	req := new(aV1.ListUserActivityModuleDistributionReq)
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	reply, err := ctl.DMS.ListUserActivityModuleDistribution(c.Request().Context(), req, currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation GET /v1/dms/statistic/user_activity/hourly_distribution UserActivity ListUserActivityHourlyDistribution
//
// List user activity hourly distribution for a date.
//
// ---
// parameters:
//   - name: stat_date
//     in: query
//     required: true
//     type: string
// responses:
//   '200':
//     description: ListUserActivityHourlyDistributionReply
//     schema:
//       "$ref": "#/definitions/ListUserActivityHourlyDistributionReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) ListUserActivityHourlyDistribution(c echo.Context) error {
	req := new(aV1.ListUserActivityHourlyDistributionReq)
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	reply, err := ctl.DMS.ListUserActivityHourlyDistribution(c.Request().Context(), req, currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation GET /v1/dms/statistic/user_activity/users UserActivity ListUserActivityUsers
//
// List user activity ranking/details.
//
// ---
// parameters:
//   - name: filter_date_from
//     in: query
//     required: true
//     type: string
//   - name: filter_date_to
//     in: query
//     required: true
//     type: string
//   - name: page_index
//     in: query
//     required: true
//     type: integer
//   - name: page_size
//     in: query
//     required: true
//     type: integer
// responses:
//   '200':
//     description: ListUserActivityUsersReply
//     schema:
//       "$ref": "#/definitions/ListUserActivityUsersReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) ListUserActivityUsers(c echo.Context) error {
	req := new(aV1.ListUserActivityUsersReq)
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	reply, err := ctl.DMS.ListUserActivityUsers(c.Request().Context(), req, currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}
