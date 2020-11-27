package models

import (
	"fmt"

	"github.com/aifece/NeteaseCloudMusicGoApi/pkg/request"
)

func (m *MusicObain) Lyric(query map[string]interface{}) map[string]interface{} {
	data := map[string]interface{}{
		"lv": -1,
		"kv": -1,
		"tv": -1,
	}
	if val, ok := query["id"]; ok {
		data["id"] = val
	}
	if val, ok := query["cookie"]; ok {
		valMapper := val.(map[string]interface{})
		valMapper["os"] = "pc"
		query["cookie"] = valMapper
	} else {
		query["cookie"] = map[string]interface{}{
			"os": "pc",
		}
	}
	fmt.Println(data)
	options := map[string]interface{}{
		"crypto": "linuxapi",
		"cookie": query["cookie"],
		"proxy":  query["proxy"],
	}
	return request.CreateRequest(
		"POST", "https://music.163.com/api/song/lyric",
		data,
		options)
}
