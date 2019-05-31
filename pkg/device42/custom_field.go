package device42

type customField struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	Notes string      `json:"notes"`
}

type customFields []customField

// GetValue retrieves a string value from a set of custom fields
func (c customFields) GetValue(key string) string {
	for _, field := range c {
		if field.Key == key {
			loc, ok := field.Value.(string)
			if !ok {
				return ""
			}
			return loc
		}
	}
	return ""
}
