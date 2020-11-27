package models

import (
	"fmt"

	"github.com/aifece/NeteaseCloudMusicGoApi/pkg/request"
)

func (m *MusicObain) Search(query map[string]interface{}) map[string]interface{} {
	data := map[string]interface{}{}
	if val, ok := query["limit"]; ok {
		data["limit"] = val
	} else {
		data["limit"] = 20
	}
	if val, ok := query["offset"]; ok {
		data["offset"] = val
	} else {
		data["offset"] = 0
	}
	if val, ok := query["type"]; ok {
		data["type"] = val
	} else {
		data["type"] = 1
	}
	if val, ok := query["keywords"]; ok {
		data["s"] = val
	} else {
		data["s"] = ""
	}

	fmt.Println(data)
	options := map[string]interface{}{
		"crypto": "weapi",
		"cookie": query["cookie"],
		"proxy":  query["proxy"],
	}
	return request.CreateRequest(
		"POST", "https://music.163.com/weapi/search/get",
		data,
		options)
}
