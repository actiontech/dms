package v1

// swagger:enum Stat
type Stat string

const (
	StatOK      Stat = "正常"
	StatDisable Stat = "被禁用"
	StatUnknown Stat = "未知"
)

type UidWithName struct {
	Uid  string `json:"uid"`
	Name string `json:"name"`
}
