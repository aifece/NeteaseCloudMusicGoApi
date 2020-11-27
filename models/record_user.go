package models

import (
	_ "fmt"

	"github.com/aifece/NeteaseCloudMusicGoApi/pkg/request"
)

func (m *MusicObain) RecordUser(query map[string]interface{}) map[string]interface{} {
	data := map[string]interface{}{
		"type":  1,
		"limit": 10,
	}
	if val, ok := query["uid"]; ok {
		data["uid"] = val
	} else {
		data["uid"] = 0
	}
	options := map[string]interface{}{
		"crypto": "weapi",
		"cookie": query["cookie"],
		"proxy":  query["proxy"],
	}

	return request.CreateRequest(
		"POST", "https://music.163.com/weapi/v1/play/record",
		data,
		options)
}
