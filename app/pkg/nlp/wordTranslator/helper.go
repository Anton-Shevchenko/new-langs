package wordTranslator

func safeGetArray(data []interface{}, idx int) ([]interface{}, bool) {
	if idx >= 0 && idx < len(data) {
		if arr, ok := data[idx].([]interface{}); ok {
			return arr, true
		}
	}
	return nil, false
}

func safeGetString(data []interface{}, idx int) (string, bool) {
	if idx >= 0 && idx < len(data) {
		if s, ok := data[idx].(string); ok {
			return s, true
		}
	}
	return "", false
}

func safeGetFloat(data []interface{}, idx int) (float64, bool) {
	if idx >= 0 && idx < len(data) {
		if f, ok := data[idx].(float64); ok {
			return f, true
		}
	}
	return 0, false
}

func SliceUnique(slice []string) []string {
	unique := make(map[string]struct{}, len(slice))
	out := make([]string, 0, len(slice))
	for _, v := range slice {
		if _, ok := unique[v]; !ok {
			unique[v] = struct{}{}
			out = append(out, v)
		}
	}
	return out
}
