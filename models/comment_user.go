package models

import (
	"fmt"
	"sync"
	"time"
	"strconv"
	"github.com/aifece/NeteaseCloudMusicGoApi/pkg/request"
)

func (m *MusicObain) CommentUser(query map[string]interface{}) map[string]interface{} {
	result := getUserCommentList(query)

	answer := map[string]interface{}{
		"status": 200,
		"body": map[string]interface{}{
			"code": 200,
			"data": map[string]interface{}{
				"comments": result,
			},
		},
		"cookie": []string{},
	}

	return answer
}

var userComment map[string][]map[string]interface{}
var userCookie map[string]interface{}
func getUserCommentList(query map[string]interface{}) []map[string]interface{} {
	uid := ""
	if val, ok := query["uid"].(string); !ok {
		return []map[string]interface{}{}
	} else {
		uid = val
	}
	if userComment == nil {
		userComment = make(map[string][]map[string]interface{})
	}
	if queryData == nil {
		queryData = make(map[string]map[string]int)
	}
	if userCookie == nil {
		userCookie = make(map[string]interface{})
	}
	if val, ok := query["cookie"]; ok {
		userCookie = val.(map[string]interface{})
	}

	userCommentList, ok := userComment[uid]
	if ok {
		return userCommentList
	} else {
		if queryData == nil {
			queryData = make(map[string]map[string]int)
		}
		queryData[uid] = map[string]int{ "page": 1, "beforeTime": 0 }
		go timeoutComment(query)
		var wg = sync.WaitGroup{}
		stopChan := make(chan bool, 1)
		wg.Add(1)
		go runGetCommentList(stopChan, &wg, query)

		return []map[string]interface{}{}
	}
}

func runGetCommentList(stopChan chan bool, wg *sync.WaitGroup, query map[string]interface{}) {
	defer wg.Done()
	uid, _ := query["uid"].(string)
	if userComment[uid] == nil {
		userComment[uid] = []map[string]interface{}{}
	}
	old_list := userComment[uid]
	for {
		more, list := getCommentList(query)
		for _, v := range list {
			old_list = append(old_list, v)
		}
		if !more {
			break
		}
	}
	userComment[uid] = old_list
}

var queryData map[string]map[string]int
func getCommentList(query map[string]interface{}) (bool, []map[string]interface{}) {
	defer func() {
        if r := recover(); r != nil {
            fmt.Printf("Play Recode Error：%s\n", r)
        }
    }()
	uid, _ := query["uid"].(string)
	userQueryData := queryData[uid]
	data := map[string]interface{}{
		"rid": query["id"],
		"offset": (userQueryData["page"] - 1) * 50,
		"limit": 50,
		"beforeTime": userQueryData["beforeTime"],
	}
	fmt.Println("Run Commnet List", uid, data)

	cookie := userCookie
	cookie["od"] = "pc"

	options := map[string]interface{}{
		"crypto": "weapi",
		"cookie": cookie,
		"proxy":  query["proxy"],
	}

	resp := request.CreateRequest(
		"POST", "https://music.163.com/api/v1/resource/comments/R_SO_4_"+fmt.Sprintf("%v", query["id"]),
		data,
		options)

	list := []map[string]interface{}{}
	body, body_ok := resp["body"].(map[string]interface{})
	if !body_ok {
		return false, list
	}
	cookie, cookie_ok := resp["cookie"].(map[string]interface{})
	if cookie_ok {
		userCookie = cookie
	}
	more, has_ok := body["more"].(bool)
	if !has_ok {
		return false, list
	}
	comments, data_ok := body["comments"].([]interface{})
	if !data_ok {
		return false, list
	}
	if more {
		userQueryData["page"] = userQueryData["page"] + 1
	}

	for _, v := range comments {
		commentItem := v.(map[string]interface{})
		time, time_ok := commentItem["time"].(int)
		if time_ok {
			userQueryData["beforeTime"] = time
		}
		if user, ok := commentItem["user"].(map[string]interface{}); ok {
			user_id, _ := user["userId"].(float64)
			if uid == strconv.Itoa(int(user_id)) {
				list = append(list, commentItem)
			} else {
				if commentId, ok := commentItem["commentId"].(float64); ok {
					floorList := getFloorCommentList(query, int(commentId))
					for _, v := range floorList {
						list = append(list, v)
					}
				}
			}
		}
	}
	queryData[uid] = userQueryData

	return more, list
}

func getFloorCommentList(query map[string]interface{}, commentId int) []map[string]interface{} {
	defer func() {
        if r := recover(); r != nil {
            fmt.Printf("Play Recode Error：%s\n", r)
        }
    }()
	options := map[string]interface{}{
		"crypto": "weapi",
		"cookie": userCookie,
		"proxy":  query["proxy"],
	}
	data := map[string]interface{}{
		"parentCommentId": commentId,
		"threadId": "R_SO_4_" + query["id"].(string),
		"limit": 20,
	}

	request_time := -1
	list := []map[string]interface{}{}
	uid, _ := query["uid"].(string)
	for {
		data["time"] = request_time
		resp := request.CreateRequest(
			"POST", "https://music.163.com/api/resource/comment/floor/get",
			data,
			options)

		body, body_ok := resp["body"].(map[string]interface{})
		if !body_ok {
			break
		}
		cookie, cookie_ok := resp["cookie"].(map[string]interface{})
		if cookie_ok {
			userCookie = cookie
		}
		data, data_ok := body["data"].(map[string]interface{})
		if !data_ok {
			break
		}
		more, has_ok := data["hasMore"].(bool)
		if !has_ok {
			break
		}
		comments, comment_ok := data["comments"].([]interface{})
		if !comment_ok {
			break
		}
		for _, v := range comments {
			commentItem := v.(map[string]interface{})
			time, time_ok := commentItem["time"].(float64)
			if time_ok {
				request_time = int(time)
			}
			if user, ok := commentItem["user"].(map[string]interface{}); ok {
				user_id, _ := user["userId"].(float64)
				if uid == strconv.Itoa(int(user_id)) {
					list = append(list, commentItem)
				}
			}
		}
		if !more {
			break
		}
	}

	return list
}

func timeoutComment(query map[string]interface{}) {
	fmt.Println("Run Query Comment Job")
	tick := time.NewTicker(60 * time.Second)
	var wg = sync.WaitGroup{}
	stopChan := make(chan bool, 1)
	for {
		wg.Add(1)
		select {
		case <-tick.C:
			go runGetCommentList(stopChan, &wg, query)
		case <-stopChan:
			goto END
		}
	}
	tick.Stop()
END:
}
