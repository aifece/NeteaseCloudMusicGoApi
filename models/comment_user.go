package models


import (
	_ "sync"
)
func (m *MusicObain) CommentUser(query map[string]interface{}) map[string]interface{} {
	uid := ""
	if val, ok := query["uid"].(string); ok {
		uid = val
	}
	id := ""
	if val, ok := query["id"].(string); ok {
		id = val
	}
	key := uid + "_" + id
	result := []interface{}{}
	if userCommentList, ok := userComment.Load(key); ok {
		for _, v := range userCommentList.(map[int]interface{}) {
			result = append(result, v)
		}
	}

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
