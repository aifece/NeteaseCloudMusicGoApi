package models

import (
	"fmt"

	"github.com/aifece/NeteaseCloudMusicGoApi/pkg/request"
)

func (m *MusicObain) CommentLike(query map[string]interface{}) map[string]interface{} {
	data := map[string]interface{}{}
	if val, ok := query["cookie"]; ok {
		valMapper := val.(map[string]interface{})
		valMapper["os"] = "pc"
		query["cookie"] = valMapper
	} else {
		query["cookie"] = map[string]interface{}{
			"os": "pc",
		}
	}
	if val, ok := query["t"]; ok && fmt.Sprintf("%v", val) == "1" {
		query["t"] = "like"
	} else {
		query["t"] = "unlike"
	}
	queryType := map[string]interface{}{
		"0": "R_SO_4_",  //  歌曲
		"1": "R_MV_5_",  //  MV
		"2": "A_PL_0_",  //  歌单
		"3": "R_AL_3_",  //  专辑
		"4": "A_DJ_1_",  //  电台,
		"5": "R_VI_62_", //  视频
		"6": "A_EV_2_",
	}[fmt.Sprintf("%v", query["type"])]
	data["threadId"] = fmt.Sprintf("%v%v", queryType, query["id"])
	data["commentId"] = query["cid"]
	if queryType == "A_EV_2_" {
		data["threadId"] = query["threadId"]
	}
	options := map[string]interface{}{
		"crypto": "weapi",
		"cookie": query["cookie"],
		"proxy":  query["proxy"],
	}
	return request.CreateRequest(
		"POST", "https://music.163.com/weapi/v1/comment/"+fmt.Sprintf("%v", query["t"]),
		data,
		options)
}
