package models

import (
	_ "fmt"

	"github.com/aifece/NeteaseCloudMusicGoApi/pkg/request"
)

func (m *MusicObain) SearchHot(query map[string]interface{}) map[string]interface{} {
	data := map[string]interface{}{
		"type": 1111,
	}

	options := map[string]interface{}{
		"crypto": "weapi",
		"ua":     "mobile",
		"cookie": query["cookie"],
		"proxy":  query["proxy"],
	}
	return request.CreateRequest(
		"POST", "https://music.163.com/weapi/search/hot",
		data,
		options)
}
