package nicovideo

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/sync/errgroup"
)

// {
// 	"video_id": "sm3000",
// 	"watch_num": 14614,
// 	"comment_num": "167",
// 	"mylist_num": 76,
// 	"title": "【MAD】最終兵器彼女",
// 	"description": "最終兵器彼女・登場キャラの死亡集を寄せ集めたようなＭＡＤ　少しばかり酷ですが、内容知ってる人は感動もできます。※オリ曲上げてくれってリクがあったが、別にあんたのために上げたんじゃないんだからね!!!オリVer＝sm1866679　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　よく見たらキリ番動画でしたｗ　　　　巡礼者の方々、乙であります！(`・ω・)ゞ",
// 	"category": "anime",
// 	"tags": "MAD アニメ 最終兵器彼女",
// 	"upload_time": "2007-03-06T12:14:17+09:00",
// 	"file_type": "flv",
// 	"length": 231,
// 	"size_high": 16558887,
// 	"size_low": 12039870
// }

var kanaConv = unicode.SpecialCase{
	// ひらがなをカタカナに変換
	unicode.CaseRange{
		Lo: 0x3041, // Lo: ぁ
		Hi: 0x3093, // Hi: ん
		Delta: [unicode.MaxCase]rune{
			0x30a1 - 0x3041, // UpperCase でカタカナに変換
			0,               // LowerCase では変換しない
			0x30a1 - 0x3041, // TitleCase でカタカナに変換
		},
	},
	// カタカナをひらがなに変換
	unicode.CaseRange{
		Lo: 0x30a1, // Lo: ァ
		Hi: 0x30f3, // Hi: ン
		Delta: [unicode.MaxCase]rune{
			0,               // UpperCase では変換しない
			0x3041 - 0x30a1, // LowerCase でひらがなに変換
			0,               // TitleCase では変換しない
		},
	},
}

type VideoFull struct {
	VideoID     string    `json:"video_id" db:"video_id"`
	WatchNum    int       `json:"watch_num" db:"watch_num"`
	CommentNum  string    `json:"comment_num" db:"comment_num"`
	MylistNum   int       `json:"mylist_num" db:"mylist_num"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	Category    string    `json:"category" db:"category"`
	Tags        string    `json:"tags" db:"tags"`
	UploadTime  time.Time `json:"upload_time" db:"upload_time"`
	FileType    string    `json:"file_type" db:"file_type"`
	Length      int       `json:"length" db:"length"`
	SizeHigh    int       `json:"size_high" db:"size_high"`
	SizeLow     int       `json:"size_low" db:"size_low"`
}

func (v *VideoFull) GetIDNum() (int, error) {
	IDStr := v.VideoID[2:]
	IDNum, err := strconv.Atoi(IDStr)
	if err != nil {
		return 0, err
	}
	return IDNum, nil
}

func (v *VideoFull) SaveToDirectory(dir string) error {
	idNum, err := v.GetIDNum()
	if err != nil {
		return err
	}
	filename := fmt.Sprintf("%04d.jsonl", idNum%1000)
	f, err := os.OpenFile(filepath.Join(dir, filename), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	return encoder.Encode(v)
}

type Video struct {
	VideoID string `json:"video_id" db:"video_id"`
	// WatchNum    int    `json:"watch_num"`
	// CommentNum  string `json:"comment_num"`
	// MylistNum   int    `json:"mylist_num"`
	// Title       string `json:"title"`
	// Description string `json:"description"`
	// Category    string `json:"category"`
	Tags       string    `json:"tags" db:"tags"`
	UploadTime time.Time `json:"upload_time" db:"upload_time"`
	// FileType    string `json:"file_type"`
	// Length      int    `json:"length"`
	// SizeHigh    int    `json:"size_high"`
	// SizeLow     int    `json:"size_low"`
}

func (v *Video) GetIDNum() (int, error) {
	IDStr := v.VideoID[2:]
	IDNum, err := strconv.Atoi(IDStr)
	if err != nil {
		return 0, err
	}
	return IDNum, nil
}

func (v *Video) IsOtomad() bool {
	return strings.Contains(v.Tags, "音MAD") || strings.Contains(v.Tags, "音mad")
}

func (v *Video) NormalizedTags() []string {
	return strings.Split(Normalize(v.Tags), " ")
}

func findJSONL(root string) []string {
	var files []string
	filepath.WalkDir(root, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".jsonl" {
			files = append(files, path)
		}
		return nil
	})

	return files
}

func readVideoFromJSONL(file string) ([]Video, error) {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var videos []Video

	decoder := json.NewDecoder(f)
	for {
		var v Video
		if err := decoder.Decode(&v); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		videos = append(videos, v)
	}
	if err != nil {
		return nil, err
	}

	return videos, nil
}

func ReadAllVideoFromDirectory(root string) ([]Video, error) {
	var videos []Video

	// ctx := context.Background()

	eg := errgroup.Group{}
	mutex := sync.Mutex{}
	// sem := semaphore.NewWeighted(3) // 最大数を3に設定

	for _, file := range findJSONL(root) {
		file := file
		// sem.Acquire(ctx, 1)
		eg.Go(func() error {
			vs, err := readVideoFromJSONL(file)
			if err != nil {
				return err
			}
			mutex.Lock()
			videos = append(videos, vs...)
			mutex.Unlock()
			// sem.Release(1)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return videos, nil
}

// func scanLastNonEmptyLine(data []byte, atEOF bool) (advance int, token []byte, err error) {
// 	// Set advance to after our last line
// 	if atEOF {
// 		advance = len(data)
// 	} else {
// 		// data[advance:] now contains a possibly incomplete line
// 		advance = bytes.LastIndexAny(data, "\n") + 1
// 	}
// 	data = data[:advance]

// 	// Remove empty lines (strip EOL chars)
// 	data = bytes.TrimRight(data, "\n")

// 	// We have no non-empty lines, so advance but do not return a token.
// 	if len(data) == 0 {
// 		return advance, nil, nil
// 	}

// 	token = data[bytes.LastIndexAny(data, "\n")+1:]
// 	return advance, token, nil
// }

func readLastVideoFromJSONL(file string) (Video, error) {
	var v Video
	f, err := os.Open(file)
	if err != nil {
		return v, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	line := ""
	for scanner.Scan() {
		line = scanner.Text()
	}
	err = json.Unmarshal([]byte(line), &v)
	if err != nil {
		return v, err
	}

	if err := scanner.Err(); err != nil {
		return v, err
	}
	return v, nil
}

func ReadLastVideoFromDirectory(root string) ([]Video, error) {
	var videos []Video

	// ctx := context.Background()

	eg := errgroup.Group{}
	mutex := sync.Mutex{}
	// sem := semaphore.NewWeighted(3) // 最大数を3に設定

	for _, file := range findJSONL(root) {
		file := file
		// sem.Acquire(ctx, 1)
		eg.Go(func() error {
			v, err := readLastVideoFromJSONL(file)
			if err != nil {
				return err
			}
			mutex.Lock()
			videos = append(videos, v)
			mutex.Unlock()
			// sem.Release(1)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return videos, nil
}

func GetCountGroupByOtomadTag(videos []Video, filter func(video Video) bool) (map[string]int, error) {
	counts := make(map[string]int)

	eg := errgroup.Group{}
	mutex := sync.Mutex{}
	for _, v := range videos {
		v := v
		eg.Go(func() error {
			if filter != nil && !filter(v) {
				return nil
			}
			tags := v.NormalizedTags()
			if !v.IsOtomad() {
				return nil
			}

			for _, tag := range tags {
				if tag == "音mad" {
					return nil
				}
				mutex.Lock()
				counts[tag]++
				mutex.Unlock()
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return counts, nil
}

func GetCountGroupByOtomadTagWithDB(year int) (map[string]int, error) {
	db, err := sqlx.Open("sqlite3", "./niconico.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	counts := make(map[string]int)

	otomadTagID := 0
	err = db.Get(&otomadTagID, "select id from tags where tag = ?", "音mad")
	if err != nil {
		return nil, err
	}

	otomadRelatedTagIDs := make([]int, 0)
	err = db.Select(&otomadRelatedTagIDs, `select distinct tag_id from video_tags where video_id in (select video_id from video_tags where tag_id = ?)`, otomadTagID)
	if err != nil {
		return nil, err
	}

	otomadRelatedTagMap := make(map[int]string, 0)
	revOtomadRelatedTagMap := make(map[string]int, 0)
	bufferSize := 1000
	otomadRelatedTagIDsBuffer := make([]int, 0)
	for _, tagID := range otomadRelatedTagIDs {
		otomadRelatedTagIDsBuffer = append(otomadRelatedTagIDsBuffer, tagID)
		if len(otomadRelatedTagIDsBuffer) == bufferSize {
			query, args, err := sqlx.In("select id, tag from tags where id in (?)", otomadRelatedTagIDsBuffer)
			if err != nil {
				return nil, err
			}
			rows, err := db.Queryx(query, args...)
			for rows.Next() {
				var id int
				var tag string
				err = rows.Scan(&id, &tag)
				if err != nil {
					return nil, err
				}
				otomadRelatedTagMap[id] = tag
				revOtomadRelatedTagMap[tag] = id
			}
			if err != nil {
				return nil, err
			}
			otomadRelatedTagIDsBuffer = make([]int, 0)
		}
	}
	if len(otomadRelatedTagIDsBuffer) > 0 {
		query, args, err := sqlx.In("select id, tag from tags where id in (?)", otomadRelatedTagIDsBuffer)
		if err != nil {
			return nil, err
		}
		rows, err := db.Queryx(query, args...)
		for rows.Next() {
			var id int
			var tag string
			err = rows.Scan(&id, &tag)
			if err != nil {
				return nil, err
			}
			otomadRelatedTagMap[id] = tag
			revOtomadRelatedTagMap[tag] = id
		}
		if err != nil {
			return nil, err
		}
	}

	otomadVideoIDs := make([]string, 0)
	err = db.Select(&otomadVideoIDs, `select video_id from video_tags where tag_id = ?`, otomadTagID)
	if err != nil {
		return nil, err
	}

	// eg := errgroup.Group{}
	// mutex := sync.Mutex{}
	for _, tID := range otomadRelatedTagIDs {
		if tID == otomadTagID {
			continue
		}
		tID2 := tID
		// eg.Go(func() error {
		videoIDs := make([]string, 0)
		if year != 0 {
			err = db.Select(&videoIDs, `select video_id from video_tags where tag_id = ?  and ? <= upload_time and upload_time < ?`,
				tID2, fmt.Sprintf("%d-01-01", year), fmt.Sprintf("%d-01-01", year+1))
		} else {
			err = db.Select(&videoIDs, `select video_id from video_tags where tag_id = ?`, tID2)
		}
		if err != nil {
			return nil, err
		}
		// mutex.Lock()
		count := len(intersection(otomadVideoIDs, videoIDs))
		counts[otomadRelatedTagMap[tID2]] = count
		// mutex.Unlock()
		// return nil
		// })
	}
	// if err := eg.Wait(); err != nil {
	// 	return nil, err
	// }
	return counts, nil
}

func GetCountGroupByTargets(videos []Video, targets []string, filter func(video Video) bool) (map[string]int, error) {
	counts := make(map[string]int)

	eg := errgroup.Group{}
	mutex := sync.Mutex{}
	for _, v := range videos {
		v := v
		eg.Go(func() error {
			if filter != nil && !filter(v) {
				return nil
			}
			tags := v.NormalizedTags()

			for _, tag := range tags {
				isTarget := false
				for _, target := range targets {
					if tag == Normalize(target) {
						isTarget = true
						break
					}
				}
				if !isTarget {
					continue
				}
				mutex.Lock()
				counts[tag]++
				mutex.Unlock()
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return counts, nil
}

func GetCountGroupByTargetsWithDB(targets []string) (map[string]int, error) {
	db, err := sqlx.Open("sqlite3", "./niconico.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	counts := make(map[string]int)
	for _, target := range targets {
		target := Normalize(target)
		var count int
		err := db.Get(&count, "select count(*) from video_tags where tag_id in (select id from tags where tag = ?)", target)
		if err != nil {
			return nil, err
		}
		counts[target] = count
	}
	return counts, nil
}

func intersection(l1, l2 []string) []string {
	s := make(map[string]struct{}, len(l1))

	// list1をmap形式に変換
	for _, data := range l1 {
		// struct{}{}何もない空のデータ
		s[data] = struct{}{}
	}

	r := make([]string, 0, len(l1))

	for _, data := range l2 {
		// mapにデータがない場合は、スキップ
		// okにはデータの存在有無true/falseで入る
		if _, ok := s[data]; !ok {
			continue
		}

		// 積集合のデータを格納
		r = append(r, data)
	}
	return r
}

func GetCountGroupByTargetsWithDBForYear(targets []string, year int) (map[string]int, error) {
	db, err := sqlx.Open("sqlite3", "./niconico.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	counts := make(map[string]int)
	eg := errgroup.Group{}
	mutex := sync.Mutex{}
	for _, target := range targets {
		target := Normalize(target)
		eg.Go(func() error {
			var count int
			err := db.Get(&count, `
		select count(*) from video_tags where tag_id in (select id from tags where tag = ?) and
				? <= upload_time and upload_time < ?`,
				target, fmt.Sprintf("%d-01-01", year), fmt.Sprintf("%d-01-01", year+1))
			if err != nil {
				return err
			}
			mutex.Lock()
			counts[target] = count
			mutex.Unlock()
			return nil
		})
		if err := eg.Wait(); err != nil {
			return nil, err
		}

	}
	return counts, nil
}
