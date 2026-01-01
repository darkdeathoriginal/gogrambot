package helpers

import "encoding/json"

func StructToMap(data any) map[string]interface{} {
	// Convert struct to map[string]interface{}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		return nil
	}
	return result
}
