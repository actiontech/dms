package dmsobject

import (
	"context"
	"fmt"
	"net/url"

	dmsV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	pkgHttp "github.com/actiontech/dms/pkg/dms-common/pkg/http"
)

// ListOpsTypes 按项目批量读取运维类型字典（只读 list；供 SQLE 等调用方一次拉全量后内存映射）。
func ListOpsTypes(ctx context.Context, dmsAddr string, req dmsV1.ListOpsTypeReq) ([]*dmsV1.OpsType, int64, error) {
	header := map[string]string{
		"Authorization": pkgHttp.DefaultDMSToken,
	}

	baseURL, err := url.Parse(fmt.Sprintf("%s%s", dmsAddr, dmsV1.GetOpsTypesRouter(req.ProjectUID)))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to parse base URL: %v", err)
	}

	query := url.Values{}
	query.Set("page_size", fmt.Sprintf("%d", req.PageSize))
	query.Set("page_index", fmt.Sprintf("%d", req.PageIndex))
	baseURL.RawQuery = query.Encode()

	reply := &dmsV1.ListOpsTypesReply{}
	if err := pkgHttp.Get(ctx, baseURL.String(), header, nil, reply); err != nil {
		return nil, 0, fmt.Errorf("failed to list ops types from %v: %v", baseURL.String(), err)
	}
	if reply.Code != 0 {
		return nil, 0, fmt.Errorf("http reply code(%v) error: %v", reply.Code, reply.Message)
	}

	return reply.Data, reply.Total, nil
}
