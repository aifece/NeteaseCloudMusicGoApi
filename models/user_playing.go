package models

import (
	"errors"
	"fmt"
	"sync"
	"time"

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

var cacheMap map[string][]interface{}
var cacheResultMap map[string]interface{}

func setResultList(uid string) {
	record_list, ok := cacheMap[uid]
	if ok {
		i := len(record_list) - 1
		new_list, ok := record_list[i].(map[int]interface{})
		if !ok {
			fmt.Println("Not found list", i)
			return
		}
		result_list := make(map[int]interface{})
		is_change := false
		for ; i > 0; i-- {
			prev_list, has_prev_list := record_list[i-1].(map[int]interface{})
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
		user_result, has_record := cacheResultMap[uid].(map[string]interface{})
		var change_list []map[string]interface{}
		change_item := map[string]interface{}{
			"change_time": check_time,
			"changes":     getSortResult(result_list),
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
			cacheResultMap[uid] = map[string]interface{}{
				"is_change":  is_change,
				"check_time": check_time,
				"result":     change_list,
			}
		} else {
			cacheResultMap[uid] = map[string]interface{}{
				"is_change":  is_change,
				"check_time": check_time,
				"result":     change_list,
			}
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
	if cacheMap == nil {
		cacheMap = make(map[string][]interface{})
	}
	if cacheResultMap == nil {
		cacheResultMap = make(map[string]interface{})
	}
	if userPlayingCookie == nil {
		userPlayingCookie = make(map[string]interface{})
	}
	if val, ok := query["cookie"]; ok {
		userPlayingCookie = val.(map[string]interface{})
	}
	uid := string(query["uid"].(string))
	_, ok := cacheMap[uid]
	if !ok {
		go timeout(query)
		var wg = sync.WaitGroup{}
		stopChan := make(chan bool, 1)
		wg.Add(1)
		go runGetList(stopChan, &wg, query)
		wg.Wait()
	}
	_, second_ok := cacheMap[uid]
	if second_ok {
		result, ok := cacheResultMap[uid].(map[string]interface{})
		if !ok {
			return nil, errors.New("还需要一些时间...")
		}
		return result, nil
	} else {
		return nil, errors.New("正在启动...")
	}
}

func runGetList(stopChan chan bool, wg *sync.WaitGroup, query map[string]interface{}) {
	defer wg.Done()

	uid := string(query["uid"].(string))
	item := getRecordList(query)
	old_list, ok := cacheMap[uid]
	if len(item) == 0 && !ok {
		stopChan <- true
		fmt.Println("Run Playing Error:", uid)
		return
	}

	if ok {
		old_list = old_list[1:]
	} else {
		old_list = make([]interface{}, 2)
		old_list[1] = item
	}
	old_list = append(old_list, item)

	cacheMap[uid] = old_list
	setResultList(uid)
}

func timeout(query map[string]interface{}) {
	fmt.Println("Run Query Job")
	tick := time.NewTicker(1 * time.Second)
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

var userPlayingCookie map[string]interface{}
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
		"proxy":  query["proxy"],
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
		userPlayingCookie = cookie
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
