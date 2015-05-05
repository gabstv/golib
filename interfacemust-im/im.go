package im

func Int(in interface{}) int {
	if v, ok := in.(int); ok {
		return v
	}
	return 0
}

func String(in interface{}) string {
	if v, ok := in.(string); ok {
		return v
	}
	return ""
}
