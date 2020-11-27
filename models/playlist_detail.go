package models

import (
	"fmt"

	"github.com/aifece/NeteaseCloudMusicGoApi/pkg/request"
)

func (m *MusicObain) PlaylistDetail(query map[string]interface{}) map[string]interface{} {
	data := map[string]interface{}{
		"n": 100000,
	}
	if val, ok := query["id"]; ok {
		data["id"] = val
	} else {
		data["id"] = ""
	}
	if val, ok := query["s"]; ok {
		data["s"] = val
	} else {
		data["s"] = 8
	}
	fmt.Println(data)
	options := map[string]interface{}{
		"crypto": "linuxapi",
		"cookie": query["cookie"],
		"proxy":  query["proxy"],
	}

	return request.CreateRequest(
		"POST", "https://music.163.com/api/v6/playlist/detail",
		data,
		options)
}
