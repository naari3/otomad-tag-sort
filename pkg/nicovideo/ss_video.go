package nicovideo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type SSVideo struct {
	StartTime      time.Time `json:"startTime"`
	MylistCounter  int       `json:"mylistCounter"`
	LengthSeconds  int       `json:"lengthSeconds"`
	CategoryTags   string    `json:"categoryTags"`
	ViewCounter    int       `json:"viewCounter"`
	ContentID      string    `json:"contentId"`
	Title          string    `json:"title"`
	CommentCounter int       `json:"commentCounter"`
	Description    string    `json:"description"`
	Tags           string    `json:"tags"`
}

func (v *SSVideo) GetIDNum() (int, error) {
	IDStr := v.ContentID[2:]
	IDNum, err := strconv.Atoi(IDStr)
	if err != nil {
		return 0, err
	}
	return IDNum, nil
}

func (v *SSVideo) ToVideo() VideoFull {
	return VideoFull{
		VideoID:     v.ContentID,
		Title:       v.Title,
		Description: v.Description,
		Tags:        v.Tags,
		WatchNum:    v.ViewCounter,
		CommentNum:  strconv.Itoa(v.CommentCounter),
		MylistNum:   v.MylistCounter,
		Length:      v.LengthSeconds,
		UploadTime:  v.StartTime,
	}
}

type SSAPIResponse struct {
	Data []SSVideo `json:"data"`
	Meta struct {
		ID         string `json:"id"`
		TotalCount int    `json:"totalCount"`
		Status     int    `json:"status"`
	} `json:"meta"`
}

// contentId,viewCounter,commentCounter,mylistCounter,title,description,categoryTags,tags,startTime,lengthSeconds

// https://api.search.nicovideo.jp/api/v2/snapshot/video/contents/search?q=&_sort=-startTime&fields=contentId,title,tags,startTime&jsonFilter=%7B%22type%22%3A%22or%22%2C%22filters%22%3A%5B%7B%22type%22%3A%22equal%22%2C%22field%22%3A%22contentId%22%2C%22value%22%3A%22sm9%22%7D%2C%7B%22type%22%3A%22equal%22%2C%22field%22%3A%22contentId%22%2C%22value%22%3A%22sm9%22%7D%2C%7B%22type%22%3A%22equal%22%2C%22field%22%3A%22contentId%22%2C%22value%22%3A%22sm9%22%7D%2C%7B%22type%22%3A%22equal%22%2C%22field%22%3A%22contentId%22%2C%22value%22%3A%22sm9%22%7D%2C%7B%22type%22%3A%22equal%22%2C%22field%22%3A%22contentId%22%2C%22value%22%3A%22sm9%22%7D%2C%7B%22type%22%3A%22equal%22%2C%22field%22%3A%22contentId%22%2C%22value%22%3A%22sm9%22%7D%2C%7B%22type%22%3A%22equal%22%2C%22field%22%3A%22contentId%22%2C%22value%22%3A%22sm9%22%7D%2C%7B%22type%22%3A%22equal%22%2C%22field%22%3A%22contentId%22%2C%22value%22%3A%22sm9%22%7D%2C%7B%22type%22%3A%22equal%22%2C%22field%22%3A%22contentId%22%2C%22value%22%3A%22sm9%22%7D%2C%7B%22type%22%3A%22equal%22%2C%22field%22%3A%22contentId%22%2C%22value%22%3A%22sm9%22%7D%5D%7D

func GetLastSSVideo() (*SSVideo, error) {
	// "https://api.search.nicovideo.jp/api/v2/snapshot/video/contents/search?q=&_sort=-startTime&fields=contentId,title"

	resp, err := http.Get("https://api.search.nicovideo.jp/api/v2/snapshot/video/contents/search?q=&_sort=-startTime&fields=contentId&_limit=1")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("StatusCode=%d", resp.StatusCode)
	}
	byteArray, _ := ioutil.ReadAll(resp.Body)
	var sSAPIResponse SSAPIResponse
	if err := json.Unmarshal(byteArray, &sSAPIResponse); err != nil {
		return nil, err
	}

	return &sSAPIResponse.Data[0], nil
}

type JSONFilter interface {
}

type JSONFilterEq struct {
	Type  string `json:"type"`
	Field string `json:"field"`
	Value string `json:"value"`
}

type JSONFilterOr struct {
	Type    string       `json:"type"`
	Filters []JSONFilter `json:"filters"`
}

func GetSSVideoFromIDs(IDs []string) ([]SSVideo, error) {
	ids := make([]JSONFilter, 0)
	for _, id := range IDs {
		ids = append(ids, JSONFilterEq{Type: "equal", Field: "contentId", Value: id})
	}
	JSONFilterOr := JSONFilterOr{
		Type:    "or",
		Filters: ids,
	}
	// marshal to json
	jsonFilter, err := json.Marshal(JSONFilterOr)
	if err != nil {
		return nil, err
	}
	j := url.QueryEscape(string(jsonFilter))
	// fmt.Printf("https://api.search.nicovideo.jp/api/v2/snapshot/video/contents/search?q=&_sort=-startTime&fields=contentId,viewCounter,commentCounter,mylistCounter,title,description,categoryTags,tags,startTime,lengthSeconds&jsonFilter=%s&_limit=%d\n", j, len(IDs))
	resp, err := http.Get(fmt.Sprintf("https://api.search.nicovideo.jp/api/v2/snapshot/video/contents/search?q=&_sort=-startTime&fields=contentId,viewCounter,commentCounter,mylistCounter,title,description,categoryTags,tags,startTime,lengthSeconds&jsonFilter=%s&_limit=%d", j, len(IDs)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("StatusCode=%d", resp.StatusCode)
	}
	byteArray, _ := ioutil.ReadAll(resp.Body)
	var sSAPIResponse SSAPIResponse
	if err := json.Unmarshal(byteArray, &sSAPIResponse); err != nil {
		return nil, err
	}
	return sSAPIResponse.Data, nil
}
