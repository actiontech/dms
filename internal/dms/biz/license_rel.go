//go:build release
// +build release

package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/license"
	"github.com/actiontech/dms/internal/pkg/locale"
	"github.com/robfig/cron/v3"
)

// license biz和data层结构复用，license本质是业务功能，且结构较为复杂，不需要在两层中做过多的结构转化
type License struct {
	UID       string    `json:"uid"  example:"zkzkzk" gorm:"primary_key"`
	CreatedAt time.Time `json:"created_at" example:"2018-10-21T16:40:23+08:00"`
	UpdatedAt time.Time `json:"updated_at" example:"2018-10-21T16:40:23+08:00"`

	WorkDurationHour int              `json:"work_duration_hour"`
	Content          *license.License `json:"content_secret" gorm:"column:content_secret;type:text"`
}

func (d *LicenseUsecase) GetLicense(ctx context.Context) (*v1.GetLicenseReply, error) {
	li, exist, err := d.repo.GetLastLicense(ctx)
	if err != nil {
		return nil, err
	}
	if !exist {
		return &v1.GetLicenseReply{License: []v1.LicenseItem{}}, nil
	}
	l, ok := li.(*License)
	if !ok {
		return nil, fmt.Errorf("get license failed")
	}

	content, err := l.Content.LicenseContent.Encode()
	if err != nil {
		return nil, err
	}

	items := generateLicenseItems(ctx, &l.Content.LicenseContent)

	items = append(items, v1.LicenseItem{
		Description: locale.Bundle.LocalizeMsgByCtx(ctx, locale.LicenseDurationOfRunning),
		Name:        "duration of running",
		Limit:       strconv.Itoa(l.WorkDurationHour / 24),
	}, v1.LicenseItem{
		Description: locale.Bundle.LocalizeMsgByCtx(ctx, locale.LicenseEstimatedMaturity),
		Name:        "estimated maturity",
		// 这个时间要展示给人看, 展示成RFC3339不够友好, 也不需要展示精确的时间, 所以展示成自定义时间格式
		Limit: time.Now().Add(time.Hour * time.Duration(l.Content.LicenseContent.Permission.WorkDurationDay*24-l.WorkDurationHour)).Format("2006-01-02"),
	})

	return &v1.GetLicenseReply{
		Content: content,
		License: items,
	}, nil

}

func (d *LicenseUsecase) GetLicenseInfo(ctx context.Context) ([]byte, error) {
	if d.clusterUsecase.IsClusterMode() {
		nodes, err := d.clusterUsecase.repo.GetClusterNodes(ctx)
		if err != nil {
			return nil, err
		}
		var clusterHardwareSigns []license.ClusterHardwareSign
		for _, node := range nodes {
			if node.ServerId != "" && node.HardwareSign != "" {
				clusterHardwareSigns = append(clusterHardwareSigns, license.ClusterHardwareSign{
					Id:   node.ServerId,
					Sign: node.HardwareSign,
				})
			}
		}

		hardwareSign, err := json.Marshal(clusterHardwareSigns)
		if err != nil {
			return nil, err
		}

		return hardwareSign, nil
	}

	hardwareSign, err := license.CollectHardwareInfo()
	if err != nil {
		return nil, license.ErrCollectLicenseInfo
	}

	return []byte(hardwareSign), nil
}

func (d *LicenseUsecase) GetLicenseUsage(ctx context.Context) (*v1.GetLicenseUsageReply, error) {
	li, exist, err := d.repo.GetLastLicense(ctx)
	if err != nil {
		return nil, err
	}

	if !exist {
		return &v1.GetLicenseUsageReply{
			Data: &v1.LicenseUsage{
				UsersUsage: v1.LicenseUsageItem{
					ResourceType:     "user",
					ResourceTypeDesc: locale.Bundle.LocalizeMsgByCtx(ctx, locale.LicenseResourceTypeUser),
					Used:             2,
					Limit:            2,
					IsLimited:        true,
				},
				DbServicesUsage: []v1.LicenseUsageItem{},
			},
		}, nil
	}

	l, ok := li.(*License)
	if !ok {
		return nil, fmt.Errorf("get license failed")
	}

	permission := l.Content.Permission
	usersTotal, err := d.userUsecase.repo.CountUsers(ctx, nil)
	if err != nil {
		return nil, err
	}

	instanceStatistics, err := d.DBService.CountDBService(ctx)
	if err != nil {
		return nil, err
	}

	dbServicesUsage := make([]v1.LicenseUsageItem, 0, len(instanceStatistics))
	var customDatabaseTypeUsage uint = 0

	for _, item := range instanceStatistics {
		if _, ok := permission.NumberOfInstanceOfEachType[item.DBType]; ok {
			usedTotal := uint(item.Count)
			limitTotal := uint(permission.NumberOfInstanceOfEachType[item.DBType].Count)
			if usedTotal > limitTotal {
				customDatabaseTypeUsage += usedTotal - limitTotal

				usedTotal = limitTotal
			}

			dbServicesUsage = append(dbServicesUsage, v1.LicenseUsageItem{
				ResourceType:     item.DBType,
				ResourceTypeDesc: item.DBType,
				Used:             usedTotal,
				Limit:            limitTotal,
				IsLimited:        true,
			})
		}
	}

	customDatabaseTypeLiteral := "custom"
	// count custom type
	if item, ok := permission.NumberOfInstanceOfEachType[customDatabaseTypeLiteral]; ok {
		dbServicesUsage = append(dbServicesUsage, v1.LicenseUsageItem{
			ResourceType:     customDatabaseTypeLiteral,
			ResourceTypeDesc: customDatabaseTypeLiteral,
			Used:             customDatabaseTypeUsage,
			Limit:            uint(item.Count),
			IsLimited:        true,
		})
	}

	return &v1.GetLicenseUsageReply{
		Data: &v1.LicenseUsage{
			UsersUsage: v1.LicenseUsageItem{
				ResourceType:     "user",
				ResourceTypeDesc: locale.Bundle.LocalizeMsgByCtx(ctx, locale.LicenseResourceTypeUser),
				Used:             uint(usersTotal),
				Limit:            uint(permission.UserCount),
				IsLimited:        true,
			},
			DbServicesUsage: dbServicesUsage,
		},
	}, nil
}

func (d *LicenseUsecase) SetLicense(ctx context.Context, data string) error {
	l := &license.License{}
	l.WorkDurationHour = 0
	err := l.Decode(data)
	if err != nil {
		return license.ErrInvalidLicense
	}

	collected, err := license.CollectHardwareInfo()
	if err != nil {
		return license.ErrCollectLicenseInfo
	}
	err = l.CheckHardwareSignIsMatch(collected)
	if err != nil {
		return err
	}

	if _, exist, err := d.repo.GetLicenseById(ctx, l.LicenseId); err != nil {
		d.log.Errorf("upload license id: %s, err: %v", l.LicenseId, err)
		return license.ErrInvalidLicense
	} else if exist {
		return license.ErrLicenseExist
	}

	err = d.repo.SaveLicense(ctx, &License{UID: l.LicenseId, Content: l, WorkDurationHour: 0})
	if err != nil {
		return fmt.Errorf("set license failed: %v", license.ErrInvalidLicense)
	}
	return nil
}

func (d *LicenseUsecase) CheckLicense(ctx context.Context, data string) (*v1.CheckLicenseReply, error) {
	l := &license.License{}
	err := l.Decode(data)
	if err != nil {
		return nil, license.ErrInvalidLicense
	}

	collected, err := license.CollectHardwareInfo()
	if err != nil {
		return nil, license.ErrCollectLicenseInfo
	}
	err = l.CheckHardwareSignIsMatch(collected)
	if err != nil {
		return nil, err
	}

	items := generateLicenseItems(ctx, &l.LicenseContent)

	return &v1.CheckLicenseReply{
		Content: string(data),
		License: items,
	}, nil

}

func generateLicenseItems(ctx context.Context, l *license.LicenseContent) []v1.LicenseItem {
	items := []v1.LicenseItem{}

	for n, i := range l.Permission.NumberOfInstanceOfEachType {
		items = append(items, v1.LicenseItem{
			Description: fmt.Sprintf(locale.Bundle.LocalizeMsgByCtx(ctx, locale.LicenseInstanceNumOfType), n),
			Name:        n,
			Limit:       strconv.Itoa(i.Count),
		})
	}

	items = append(items, v1.LicenseItem{
		Description: locale.Bundle.LocalizeMsgByCtx(ctx, locale.LicenseUserNum),
		Name:        "user",
		Limit:       strconv.Itoa(l.Permission.UserCount),
	})

	if l.HardwareSign != "" {
		items = append(items, v1.LicenseItem{
			Description: locale.Bundle.LocalizeMsgByCtx(ctx, locale.LicenseMachineInfo),
			Name:        "info",
			Limit:       l.HardwareSign,
		})
	}
	if len(l.ClusterHardwareSigns) > 0 {
		for _, s := range l.ClusterHardwareSigns {
			items = append(items, v1.LicenseItem{
				Description: fmt.Sprintf(locale.Bundle.LocalizeMsgByCtx(ctx, locale.LicenseMachineInfoOfNode), s.Id),
				Name:        fmt.Sprintf("node_%s_info", s.Id),
				Limit:       s.Sign,
			})
		}
	}
	items = append(items, []v1.LicenseItem{
		{
			Description: locale.Bundle.LocalizeMsgByCtx(ctx, locale.LicenseDmsVersion),
			Name:        "version",
			Limit:       l.Permission.Version,
		}, {
			Description: locale.Bundle.LocalizeMsgByCtx(ctx, locale.LicenseAuthorizedDurationDay),
			Name:        "work duration day",
			Limit:       strconv.Itoa(l.Permission.WorkDurationDay),
		},
	}...)

	return items
}

func (d *LicenseUsecase) Check(ctx context.Context) (*v1.GetLicenseReply, error) {
	license, exist, err := d.repo.GetLastLicense(ctx)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	l, ok := license.(*License)
	if !ok {
		return nil, fmt.Errorf("get license failed")
	}
	content, err := l.Content.LicenseContent.Encode()
	if err != nil {
		return nil, err
	}

	items := generateLicenseItems(ctx, &l.Content.LicenseContent)

	items = append(items, v1.LicenseItem{
		Description: locale.Bundle.LocalizeMsgByCtx(ctx, locale.LicenseDurationOfRunning),
		Name:        "duration of running",
		Limit:       strconv.Itoa(l.WorkDurationHour / 24),
	}, v1.LicenseItem{
		Description: locale.Bundle.LocalizeMsgByCtx(ctx, locale.LicenseEstimatedMaturity),
		Name:        "estimated maturity",
		// 这个时间要展示给人看, 展示成RFC3339不够友好, 也不需要展示精确的时间, 所以展示成自定义时间格式
		Limit: time.Now().Add(time.Hour * time.Duration(l.Content.LicenseContent.Permission.WorkDurationDay*24-l.WorkDurationHour)).Format("2006-01-02"),
	})

	return &v1.GetLicenseReply{
		Content: content,
		License: items,
	}, nil

}

func (d *LicenseUsecase) GetLicenseInner(ctx context.Context) (*License, bool, error) {
	license, exist, err := d.repo.GetLastLicense(ctx)
	if err != nil {
		return nil, false, err
	}
	if !exist {
		return nil, false, fmt.Errorf("license is not exist")
	}
	l, ok := license.(*License)
	if !ok {
		return nil, false, fmt.Errorf("get license failed")
	}
	return l, true, nil
}

func (d *LicenseUsecase) getLicenseInner(ctx context.Context) (*license.License, error) {
	license, exist, err := d.repo.GetLastLicense(ctx)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("license is not exist")
	}
	l, ok := license.(*License)
	if !ok {
		return nil, fmt.Errorf("get license failed")
	}
	if l.Content == nil {
		return nil, fmt.Errorf("license is empty")
	}
	return l.Content, nil
}

func (d *LicenseUsecase) CheckCanCreateUser(ctx context.Context) (bool, error) {
	count, err := d.userUsecase.repo.CountUsers(ctx, nil)
	if err != nil {
		return true, err
	}
	l, err := d.getLicenseInner(ctx)
	if err != nil {
		return true, err
	}
	err = l.CheckCanCreateUser(count)
	if err != nil {
		return true, err
	}
	return true, nil
}

func (d *LicenseUsecase) CheckCanCreateInstance(ctx context.Context, dbType string) (bool, error) {
	l, err := d.getLicenseInner(ctx)
	if err != nil {
		return true, err
	}
	counts, err := d.DBService.CountDBService(ctx)
	if err != nil {
		return true, err
	}

	usage := make(license.LimitOfEachType, len(counts))
	for _, v := range counts {
		usage[v.DBType] = license.LimitOfType{
			DBType: v.DBType,
			Count:  int(v.Count),
		}
	}
	err = l.CheckCanCreateInstance(dbType, usage)
	if err != nil {
		return true, err
	}
	return false, nil
}

func (d *LicenseUsecase) initial() {
	if d.cron == nil {
		d.cron = cron.New()
	}

	if _, err := d.cron.AddFunc("@hourly", func() {
		// 集群模式，只有leader节点需要执行定时任务
		if d.clusterUsecase.IsClusterMode() && !d.clusterUsecase.IsLeader() {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		li, exist, err := d.repo.GetLastLicense(ctx)
		if err != nil {
			d.log.Errorf("find license err: %v", err)
			return
		}
		if !exist {
			return
		}

		l, _ := li.(*License)
		l.WorkDurationHour++
		l.Content.WorkDurationHour++

		if err = d.repo.SaveLicense(ctx, l); err != nil {
			d.log.Errorf("update license err: %v", err)
		}
	}); err != nil {
		d.log.Errorf("license cron job err: %v", err)
	}

	d.cron.Start()
}
