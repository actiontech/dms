package strconv

import (
	"fmt"
	"strconv"
)

// 由于前端处理int64类型会出现精度丢失，所以与前端交互的uid全部使用string类型
// uid统一用用该方法来做转换
func ParseUid(uidStr string) (int64, error) {
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if nil != err {
		return 0, err
	}
	return uid, nil
}

func UidArrStr2Int(strArr []string) (res []int64, err error) {
	res = make([]int64, len(strArr))
	for index, val := range strArr {
		res[index], err = ParseUid(val)
		if nil != err {
			return nil, err
		}
	}
	return res, nil
}

func UidInt2Str(uid int64) string {
	return fmt.Sprintf("%d", uid)
}

func StringInArr(str string, arr []string) bool {
	for i := range arr {
		if arr[i] == str {
			return true
		}
	}
	return false
}
