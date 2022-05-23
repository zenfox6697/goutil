package api

func FuckPage(obj []interface{}, limit, offset int) map[string]interface{} {
	var res []interface{}
	for k, v := range obj {
		if k >= offset && k < limit+offset {
			res = append(res, v)
		}
	}
	if len(res) > 0 {
		return map[string]interface{}{
			"count": len(res),
			"obj":   res,
		}
	}
	return map[string]interface{}{
		"count": 0,
		"obj":   []interface{}{},
	}
}
