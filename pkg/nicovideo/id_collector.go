package nicovideo

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type IDCollector struct {
	From        int
	CurrentPage int
	IsLastPage  bool
	IDs         []string
	Idx         int
}

func NewIDCollector(from int) (*IDCollector, error) {
	firstPage := int(math.Ceil(float64(from) / 200.0))
	result, err := getIDsFromPage(firstPage)
	if err != nil {
		return nil, err
	}
	idx := 0
	// skip idx until larger than from id
	for i, id := range result.IDs {
		idNum, err := strconv.Atoi(id[len("sm"):])
		if err != nil {
			return nil, err
		}
		if idNum >= from {
			idx = i
			break
		}
	}
	idCollector := IDCollector{
		From:        from,
		CurrentPage: firstPage,
		IsLastPage:  !result.HasNextPage,
		IDs:         result.IDs,
		Idx:         idx,
	}
	return &idCollector, nil
}

func (c *IDCollector) HasNext() bool {
	leachToLastItem := c.Idx >= len(c.IDs)
	if leachToLastItem {
		if c.IsLastPage {
			return false
		} else {
			c.CurrentPage++
			fmt.Println("Get more contents from next page", c.CurrentPage)
			result, err := getIDsFromPage(c.CurrentPage)
			if err != nil {
				return false
			}
			if len(result.IDs) == 0 {
				fmt.Println("Probably in the process of creating a page, check one more page")
				c.CurrentPage++
				result, err = getIDsFromPage(c.CurrentPage)
				if err != nil {
					return false
				}
				if len(result.IDs) == 0 {
					return false
				}
			}
			fmt.Println("result ids: ", result.IDs[0], "...", result.IDs[len(result.IDs)-1])
			// fmt.Println("result HasNextPage: ", result.HasNextPage)
			c.IDs = result.IDs
			c.IsLastPage = !result.HasNextPage
			c.Idx = 0
			return true
		}
	} else {
		// fmt.Println("valid", c.Idx, c.IDs[c.Idx])
		return true
	}
}

func (c *IDCollector) Next() string {
	if c.HasNext() {
		c.Idx++
		// fmt.Println(c.IDs[c.Idx-1])
		return c.IDs[c.Idx-1]
	} else {
		return ""
	}
}

type VideoCatalogResult struct {
	IDs         []string
	HasPrevPage bool
	HasNextPage bool
}

func getIDsFromPage(page int) (*VideoCatalogResult, error) {
	fmt.Println("getIDsFromPage:", page)
	var result VideoCatalogResult
	resp, err := http.Get("https://www.nicovideo.jp/video_catalog/3/" + strconv.Itoa(page))
	time.Sleep(time.Millisecond * 100)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("StatusCode=%d", resp.StatusCode)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0)
	doc.Find("body > div.BaseLayout-main > div > ul > li > a").Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if ok {
			id := href[len("https://www.nicovideo.jp/watch/"):]
			if id != "" {
				ids = append(ids, id)
			}
		}
	})
	sort.SliceStable(ids, func(i, j int) bool {
		a, err := strconv.Atoi(ids[i][len("sm"):])
		if err != nil {
			panic(err)
		}
		b, err := strconv.Atoi(ids[j][len("sm"):])
		if err != nil {
			panic(err)
		}
		return a < b
	})
	result.IDs = ids

	// 前のページがあるかどうか
	prevPage := doc.Find("span.VideoCatalogPage-prevPage > a")
	if prevPage.Length() > 0 {
		result.HasPrevPage = true
	}
	// 次のページがあるかどうか
	nextPage := doc.Find("span.VideoCatalogPage-nextPage > a")
	if nextPage.Length() > 0 {
		result.HasNextPage = true
	}

	return &result, nil
}
