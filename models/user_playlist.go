package models

import (
	_ "fmt"

	"github.com/aifece/NeteaseCloudMusicGoApi/pkg/request"
)

func (m *MusicObain) UserPlaylist(query map[string]interface{}) map[string]interface{} {
	data := map[string]interface{}{
		"includeVideo": true,
	}
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
	if val, ok := query["uid"]; ok {
		data["uid"] = val
	}
	options := map[string]interface{}{
		"crypto": "weapi",
		"cookie": query["cookie"],
		"proxy":  query["proxy"],
	}

	return request.CreateRequest(
		"POST", "https://music.163.com/api/user/playlist",
		data,
		options)
}
