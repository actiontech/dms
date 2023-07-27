package strconv

func NvlString(param *string) string {
	if param != nil {
		return *param
	}
	return ""
}
