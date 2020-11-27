package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/aifece/NeteaseCloudMusicGoApi/pkg/request"
)

func (m *MusicObain) SongDetail(query map[string]interface{}) map[string]interface{} {
	data := map[string]interface{}{}
	if val, ok := query["ids"]; ok {
		ids := strings.Split(val.(string), ",")
		enable_ids := make([]int, 0)
		data_c := make([]map[string]int, 0)
		for _, v := range ids {
			tmp_v, _ := strconv.Atoi(v)
			if tmp_v > 0 {
				enable_ids = append(enable_ids, tmp_v)
				data_c = append(data_c, map[string]int{
					"id": tmp_v,
				})
			}
		}

		str_c, _ := json.Marshal(data_c)
		str_ids, _ := json.Marshal(enable_ids)
		data["c"] = string(str_c)
		data["ids"] = string(str_ids)
	}
	fmt.Println(data)
	options := map[string]interface{}{
		"crypto": "weapi",
		"cookie": query["cookie"],
		"proxy":  query["proxy"],
	}
	return request.CreateRequest(
		"POST", "https://music.163.com/weapi/v3/song/detail",
		data,
		options)
}
