//go:build release
// +build release

package biz

import (
	"context"
	"fmt"
	"strconv"
	"time"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/license"
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

	items := generateLicenseItems(&l.Content.LicenseContent)

	items = append(items, v1.LicenseItem{
		Description: "已运行时长(天)",
		Name:        "duration of running",
		Limit:       strconv.Itoa(l.WorkDurationHour / 24),
	}, v1.LicenseItem{
		Description: "预计到期时间",
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
	hardwareSign, err := license.CollectHardwareInfo()
	if err != nil {
		return nil, license.ErrCollectLicenseInfo
	}
	return []byte(hardwareSign), nil
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

	items := generateLicenseItems(&l.LicenseContent)

	return &v1.CheckLicenseReply{
		Content: string(data),
		License: items,
	}, nil

}

func generateLicenseItems(l *license.LicenseContent) []v1.LicenseItem {
	items := []v1.LicenseItem{}

	for n, i := range l.Permission.NumberOfInstanceOfEachType {
		items = append(items, v1.LicenseItem{
			Description: fmt.Sprintf("[%v]类型实例数", n),
			Name:        n,
			Limit:       strconv.Itoa(i.Count),
		})
	}

	items = append(items, v1.LicenseItem{
		Description: "用户数",
		Name:        "user",
		Limit:       strconv.Itoa(l.Permission.UserCount),
	})

	if l.HardwareSign != "" {
		items = append(items, v1.LicenseItem{
			Description: "机器信息",
			Name:        "info",
			Limit:       l.HardwareSign,
		})
	}
	if len(l.ClusterHardwareSigns) > 0 {
		for _, s := range l.ClusterHardwareSigns {
			items = append(items, v1.LicenseItem{
				Description: fmt.Sprintf("节点[%s]机器信息", s.Id),
				Name:        fmt.Sprintf("node_%s_info", s.Id),
				Limit:       s.Sign,
			})
		}
	}
	items = append(items, []v1.LicenseItem{
		{
			Description: "DMS版本",
			Name:        "version",
			Limit:       l.Permission.Version,
		}, {
			Description: "授权运行时长(天)",
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

	items := generateLicenseItems(&l.Content.LicenseContent)

	items = append(items, v1.LicenseItem{
		Description: "已运行时长(天)",
		Name:        "duration of running",
		Limit:       strconv.Itoa(l.WorkDurationHour / 24),
	}, v1.LicenseItem{
		Description: "预计到期时间",
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
	_, count, err := d.userUsecase.ListUser(ctx, &ListUsersOption{PageNumber: 1, LimitPerPage: 1})
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
