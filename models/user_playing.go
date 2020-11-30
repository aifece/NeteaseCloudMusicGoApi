package models

import (
	"errors"
	"fmt"
	"sync"
	"time"
	"strconv"
	"github.com/aifece/NeteaseCloudMusicGoApi/pkg/request"
)

func (m *MusicObain) UserPlaying(query map[string]interface{}) map[string]interface{} {
	result, _ := getUserPlayList(query)

	answer := map[string]interface{}{
		"status": 200,
		"body": map[string]interface{}{
			"code": 200,
			"data": result,
		},
		"cookie": []string{},
	}

	return answer
}

var cacheMap sync.Map
var cacheResultMap sync.Map
var userPlayingCookie sync.Map
var userPlayingProxy sync.Map
var userComment sync.Map
var queryData sync.Map
func setResultList(uid string) {
	if val, ok := cacheMap.Load(uid); ok {
		record_list := val.([]map[int]interface{})
		i := len(record_list) - 1
		new_list := record_list[i]
		if len(new_list) == 0 {
			fmt.Println("Not found list", i)
			return
		}
		result_list := make(map[int]interface{})
		is_change := false
		for ; i > 0; i-- {
			prev_list := record_list[i-1]
			has_prev_list := len(prev_list) > 0
			for id, v := range new_list {
				new_item := v.(map[string]interface{})
				new_index := new_item["index"].(int)
				score := 100 - new_index
				result_item := make(map[string]interface{})
				if val, ok := result_list[id]; !ok {
					result_item = new_item
				} else {
					result_item = val.(map[string]interface{})
				}
				if val, ok := result_item["score"]; ok {
					score = val.(int)
				}
				if !has_prev_list {
					is_change = true
					result_item["score"] = score
					result_list[id] = result_item
					continue
				}

				tmp_change := false
				if prev_item, ok := prev_list[id].(map[string]interface{}); ok {
					prev_index, _ := prev_item["index"].(int)
					if new_index < prev_index {
						score = score + (prev_index - new_index) * 128
						is_change = true
						tmp_change = true
					}
				} else {
					score = score + 1994
					is_change = true
					tmp_change = true
				}
				if tmp_change {
					result_item["score"] = score
					result_list[id] = result_item
				}
			}
		}
		now := time.Now()
		check_time := fmt.Sprintf("%d-%d-%d %d:%d:%d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())

		user_result := map[string]interface{}{}
		has_record := false
		if val, ok := cacheResultMap.Load(uid); ok {
			user_result = val.(map[string]interface{})
			has_record = true
		}

		var change_list []map[string]interface{}
		sort_result_list := getSortResult(result_list)
		change_item := map[string]interface{}{
			"change_time": check_time,
			"changes":     sort_result_list,
		}
		if has_record {
			change_list = user_result["result"].([]map[string]interface{})
		} else {
			change_list = make([]map[string]interface{}, 0)
		}
		if is_change {
			change_length := len(change_list)
			max := 99
			min := 0
			if change_length < max {
				max = change_length
			}
			if change_length > max {
				min = 1
			}
			change_list = append(change_list[min:max], change_item)
			cacheResultMap.Store(uid, map[string]interface{}{
				"is_change":  is_change,
				"check_time": check_time,
				"result":     change_list,
			})
			// 获取Ta的评论
			for _, v := range sort_result_list {
				if id, ok := v["id"].(int); ok {
					go getUserCommentList(strconv.Itoa(id), uid)
				}
			}
		} else {
			cacheResultMap.Store(uid, map[string]interface{}{
				"is_change":  is_change,
				"check_time": check_time,
				"result":     change_list,
			})
		}
	}
}

func getSortResult(list map[int]interface{}) []map[string]interface{} {
	length := len(list)
	list_arr := make([]map[string]interface{}, length)
	i := 0
	for _, v := range list {
		tmp := v.(map[string]interface{})
		list_arr[i] = tmp
		i++
	}
	max := 10
	if max > length {
		max = length
	}

	for i := 0; i < length; i++ {
		curr_score := list_arr[i]["score"].(int)
		for j := i + 1; j < length; j++ {
			next_socre := list_arr[j]["score"].(int)
			if curr_score < next_socre {
				tmp_item := list_arr[i]
				list_arr[i] = list_arr[j]
				list_arr[j] = tmp_item
				curr_score = next_socre
			}
		}
	}

	return list_arr[:max]
}

func getUserPlayList(query map[string]interface{}) (map[string]interface{}, error) {
	if val, ok := query["uid"]; ok {
		query["uid"] = val
	}
	if val, ok := query["cookie"]; ok {
		for k, v := range val.(map[string]interface{}) {
			userPlayingCookie.Store(k, v)
		}
	}
	userPlayingCookie.Store("od", "pc")
	if val, ok := query["proxy"]; ok {
		for k, v := range val.(map[string]interface{}) {
			userPlayingProxy.Store(k, v)
		}
	}
	uid := string(query["uid"].(string))
	_, ok := cacheMap.Load(uid)
	if !ok {
		go timeout(query)
		var wg = sync.WaitGroup{}
		stopChan := make(chan bool, 1)
		wg.Add(1)
		go runGetList(stopChan, &wg, query)
		wg.Wait()
	}
	_, second_ok := cacheMap.Load(uid)
	if second_ok {
		if result, ok := cacheResultMap.Load(uid); ok {
			return result.(map[string]interface{}), nil
		} else {
			return nil, errors.New("还需要一些时间...")
		}
	} else {
		return nil, errors.New("正在启动...")
	}
}

func runGetList(stopChan chan bool, wg *sync.WaitGroup, query map[string]interface{}) {
	defer wg.Done()

	uid := string(query["uid"].(string))
	item := getRecordList(query)
	list, _ := cacheMap.Load(uid)
	old_list, ok := list.([]map[int]interface{})
	if len(item) == 0 && !ok {
		stopChan <- true
		fmt.Println("Run Playing Error:", uid)
		return
	}

	if ok {
		old_list = old_list[1:]
	} else {
		old_list = make([]map[int]interface{}, 2)
		old_list[1] = item
	}
	old_list = append(old_list, item)

	cacheMap.Store(uid, old_list)
	setResultList(uid)
}

func getRecordList(query map[string]interface{}) map[int]interface{} {
	defer func() {
        if r := recover(); r != nil {
            fmt.Printf("Play Recode Error：%s\n", r)
        }
    }()
	list := map[int]interface{}{}
	request_data := map[string]interface{}{
		"type":      1,
		"limit":     100,
		"showCount": false,
	}
	uid := ""
	if val, ok := query["uid"]; ok {
		uid = val.(string)
	}
	request_data["uid"] = uid

	options := map[string]interface{}{
		"crypto": "weapi",
		"cookie": userPlayingCookie,
		"proxy":  userPlayingProxy,
	}
	resp := request.CreateRequest(
		"POST", "https://music.163.com/weapi/v1/play/record",
		request_data,
		options)
	body, body_ok := resp["body"].(map[string]interface{})
	if !body_ok {
		return list
	}
	cookie, cookie_ok := resp["cookie"].(map[string]interface{})
	if cookie_ok {
		for k, v := range cookie {
			userPlayingCookie.Store(k, v)
		}
	}
	weekData, data_ok := body["weekData"].([]interface{})
	if !data_ok {
		return list
	}

	for i, v := range weekData {
		v_m := v.(map[string]interface{})
		song := v_m["song"].(map[string]interface{})
		song_al := song["al"].(map[string]interface{})
		id := int(song["id"].(float64))
		source_id := int(song_al["id"].(float64))
		list[id] = map[string]interface{}{
			"id":        id,
			"source_id": source_id,
			"index":     i,
			"name":      song["name"].(string),
		}
	}

	return list
}


func getUserCommentList(id string, uid string) {
	key := uid + "_" + id
	_, ok := userComment.Load(key)
	if !ok {
		queryData.Store(key, map[string]int{ "page": 1, "beforeTime": 0, "only_one": 0 })
		go timeoutComment(id, uid)
		var wg = sync.WaitGroup{}
		wg.Add(1)
		go runGetCommentList(&wg, id, uid)
	}
}

func runGetCommentList(wg *sync.WaitGroup, id string, uid string) {
	defer wg.Done()
	defer func() {
        if r := recover(); r != nil {
            fmt.Printf("Run Get Comment Go Error：%s\n", r)
        }
    }()

	key := uid + "_" + id
	old_list := map[int]interface{}{}
	if val, ok := userComment.Load(key); ok {
		old_list = val.(map[int]interface{})
	}
	for {
		more, list := getCommentList(id, uid)
		if len(list) > 0 {
			fmt.Println("User Has Comment: ", key)
		}
		for _, v := range list {
			if commentId, ok := v["commentId"].(float64); ok {
				old_list[int(commentId)] = v
				userComment.Store(key, old_list)
			}
		}
		if !more {
			break
		}
	}
	userComment.Store(key, old_list)
}

func getCommentList(id string, uid string) (bool, []map[string]interface{}) {
	defer func() {
        if r := recover(); r != nil {
            fmt.Printf("Comment Recode Error：%s\n", r)
        }
    }()
    key := uid + "_" + id
    userQueryData := map[string]int{}
	if val, ok := queryData.Load(key); ok {
		userQueryData = val.(map[string]int)
	}
	data := map[string]interface{}{
		"rid": id,
		"offset": (userQueryData["page"] - 1) * 50,
		"limit": 50,
		"beforeTime": userQueryData["beforeTime"],
	}

	list := []map[string]interface{}{}
	if userQueryData["page"] > 20 {
		userQueryData["page"] = 1
		userQueryData["only_one"] = 1
		userQueryData["beforeTime"] = 0
		queryData.Store(key, userQueryData)
		return false, list
	}

	options := map[string]interface{}{
		"crypto": "weapi",
		"cookie": userPlayingCookie,
		"proxy":  userPlayingProxy,
	}

	resp := request.CreateRequest(
		"POST", "https://music.163.com/api/v1/resource/comments/R_SO_4_" + id,
		data,
		options)

	body, body_ok := resp["body"].(map[string]interface{})
	if !body_ok {
		return false, list
	}
	cookie, cookie_ok := resp["cookie"].(map[string]interface{})
	if cookie_ok {
		for k, v := range cookie {
			userPlayingCookie.Store(k, v)
		}
	}
	more, has_ok := body["more"].(bool)
	if !has_ok {
		return false, list
	}
	if userQueryData["only_one"] == 1 {
		more = false
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
		if time, ok := commentItem["time"].(float64); ok {
			userQueryData["beforeTime"] = int(time)
		}
		if user, ok := commentItem["user"].(map[string]interface{}); ok {
			user_id, _ := user["userId"].(float64)
			if uid == strconv.Itoa(int(user_id)) {
				list = append(list, commentItem)
			} else {
				if commentId, ok := commentItem["commentId"].(float64); ok {
					floorList := getFloorCommentList(id, uid, int(commentId))
					for _, v := range floorList {
						list = append(list, v)
					}
				}
			}
		}
	}
	queryData.Store(key, userQueryData)

	return more, list
}

func getFloorCommentList(id string, uid string, commentId int) []map[string]interface{} {
	defer func() {
        if r := recover(); r != nil {
            fmt.Printf("Floor Comment Recode Error：%s\n", r)
        }
    }()
	options := map[string]interface{}{
		"crypto": "weapi",
		"cookie": userPlayingCookie,
		"proxy":  userPlayingProxy,
	}
	data := map[string]interface{}{
		"parentCommentId": commentId,
		"threadId": "R_SO_4_" + id,
		"limit": 20,
	}

	request_time := -1
	list := []map[string]interface{}{}
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
			for k, v := range cookie {
				userPlayingCookie.Store(k, v)
			}
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

func timeout(query map[string]interface{}) {
	fmt.Println("Run Query Job")
	tick := time.NewTicker(30 * time.Second)
	var wg = sync.WaitGroup{}
	stopChan := make(chan bool, 1)
	for {
		wg.Add(1)
		select {
		case <-tick.C:
			go runGetList(stopChan, &wg, query)
		case <-stopChan:
			goto END
		}
	}
	tick.Stop()
END:
}

func timeoutComment(id string, uid string) {
	fmt.Println("Run Query Comment Job")
	tick := time.NewTicker(60 * time.Second)
	var wg = sync.WaitGroup{}
	for {
		wg.Add(1)
		select {
		case <-tick.C:
			go runGetCommentList(&wg, id, uid)
		}
	}
}
