package models

import (
	"fmt"

	"github.com/aifece/NeteaseCloudMusicGoApi/pkg/request"
)

func (m *MusicObain) SongUrl(query map[string]interface{}) map[string]interface{} {
	data := map[string]interface{}{
		"br": 999000,
	}
	if val, ok := query["id"]; ok {
		data["ids"] = "[" + val.(string) + "]"
	}
	fmt.Println(data)
	options := map[string]interface{}{
		"crypto": "eapi",
		"cookie": query["cookie"],
		"proxy":  query["proxy"],
		"url":    "/api/song/enhance/player/url",
	}

	return request.CreateRequest(
		"POST", "https://interface3.music.163.com/eapi/song/enhance/player/url",
		data,
		options)
}
