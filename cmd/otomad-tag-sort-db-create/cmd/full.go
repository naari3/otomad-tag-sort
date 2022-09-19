package cmd

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/naari3/otomad-tag-sort/pkg/nicovideo"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(fullCmd)
}

var fullCmd = &cobra.Command{
	Use:   "full",
	Short: "Create full database",
	Long:  `Create full database`,
	Run: func(cmd *cobra.Command, args []string) {
		runFull()
	},
}

var schema = `
create table videos (
	video_id text not null primary key,
	tags text,
	upload_time text,
	upload_year integer
);
create index videos_upload_time on videos(upload_time);

create table video_tags (
	id integer not null primary key autoincrement,
	video_id text not null,
	tag_id integer not null
);
create index video_tags_video_id on video_tags(video_id);
create index video_tags_tag_id on video_tags(tag_id);

create table tags (
	id integer not null primary key autoincrement,
	tag text not null unique
);
create index tags_tag on tags(tag);
`

type Video struct {
	VideoID    string    `db:"video_id"`
	Tags       string    `db:"tags"`
	UploadTime time.Time `db:"upload_time"`
	UploadYear int       `db:"upload_year"`
}

func (v *Video) NormalizedTags() []string {
	return strings.Split(nicovideo.Normalize(v.Tags), " ")
}

func runFull() error {
	if err := os.Remove("niconico.db"); err != nil {
		log.Fatal(err)
	}
	db, err := sqlx.Open("sqlite3", "./niconico.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(schema)
	if err != nil {
		log.Printf("%q: %s\n", err, schema)
	}

	_, err = db.Exec(`delete from videos;`)
	if err != nil {
		log.Printf("%q: %s\n", err, schema)
	}

	videos, err := nicovideo.ReadAllVideoFromDirectory("jsonl")
	if err != nil {
		panic(err)
	}
	log.Println("Collected all videos:", len(videos))

	append_videos, err := nicovideo.ReadAllVideoFromDirectory("jsonl_append")
	if err != nil {
		panic(err)
	}
	log.Println("Collected all append_videos:", len(append_videos))

	videos = append(videos, append_videos...)
	unique_video_map := make(map[string]Video)
	for _, v := range videos {
		if _, ok := unique_video_map[v.VideoID]; ok {
			continue
		}
		unique_video_map[v.VideoID] = Video{
			VideoID:    v.VideoID,
			Tags:       v.Tags,
			UploadTime: v.UploadTime,
			UploadYear: v.UploadTime.Year(),
		}
	}
	log.Println("Collected all unique videos:", len(unique_video_map))
	tagID := 0
	tagIDMap := make(map[string]int)
	for _, v := range unique_video_map {
		for _, t := range v.NormalizedTags() {
			if _, ok := tagIDMap[t]; !ok {
				tagIDMap[t] = tagID
				tagID++
			}
		}
	}
	log.Println("Collected all unique tags:", len(tagIDMap))

	type Tag struct {
		ID  int    `db:"id"`
		Tag string `db:"tag"`
	}
	tagsBufSize := 400
	bufTags := make([]Tag, 0, tagsBufSize)
	tagCount := 0
	for t := range tagIDMap {
		if tagCount > tagsBufSize && tagCount%100000 == 0 {
			log.Println("Processing", tagCount, "tags")
		}
		bufTags = append(bufTags, Tag{ID: tagIDMap[t], Tag: t})
		if len(bufTags) == tagsBufSize {
			_, err = db.NamedExec(`insert into tags(id, tag) values(:id, :tag)`, bufTags)
			if err != nil {
				log.Fatal(err)
			}
			bufTags = bufTags[:0]
		}
		tagCount++
	}
	if len(bufTags) > 0 {
		_, err = db.NamedExec(`insert into tags(id, tag) values(:id, :tag)`, bufTags)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Println("Inserted all tags")

	videoBufSize := 300
	bufVideo := make([]Video, 0, videoBufSize)
	videoCount := 0
	videoTagIDsMap := make(map[string][]int)
	for _, v := range unique_video_map {
		if videoCount > 0 && videoCount%100000 == 0 {
			log.Println("Processing", videoCount, "videos")
		}
		bufVideo = append(bufVideo, v)
		if _, ok := videoTagIDsMap[v.VideoID]; !ok {
			videoTagIDsMap[v.VideoID] = make([]int, 0)
		}
		tagIDs := make([]int, 0, len(v.NormalizedTags()))
		for _, t := range v.NormalizedTags() {
			tagIDs = append(tagIDs, tagIDMap[t])
		}
		videoTagIDsMap[v.VideoID] = tagIDs

		if len(bufVideo) == videoBufSize {
			_, err = db.NamedExec(`insert into videos(video_id, tags, upload_time, upload_year) values(:video_id, :tags, :upload_time, :upload_year)`, bufVideo)
			if err != nil {
				log.Fatal(err)
			}
			bufVideo = bufVideo[:0]
		}
		videoCount++
	}
	if len(bufVideo) > 0 {
		_, err = db.NamedExec(`insert into videos(video_id, tags, upload_time) values(:video_id, :tags, :upload_time)`, bufVideo)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Println("Inserted all videos")
	log.Println("Collected all video_tags:", len(videoTagIDsMap))

	type VideoTag struct {
		ID      int    `db:"id"`
		VideoID string `db:"video_id"`
		TagID   int    `db:"tag_id"`
	}

	videoTagBufSize := 400
	bufVideoTag := make([]VideoTag, 0, videoTagBufSize)
	videoTagCount := 0
	for videoID, tagIDs := range videoTagIDsMap {
		if videoTagCount > 0 && videoTagCount%100000 == 0 {
			log.Println("Processing", videoTagCount, "video_tags")
		}
		for _, tagID := range tagIDs {
			bufVideoTag = append(bufVideoTag, VideoTag{
				VideoID: videoID,
				TagID:   tagID,
			})
			if len(bufVideoTag) == videoTagBufSize {
				_, err = db.NamedExec(`insert into video_tags(video_id, tag_id) values(:video_id, :tag_id)`, bufVideoTag)
				if err != nil {
					log.Println(bufVideoTag[0], "...", bufVideoTag[len(bufVideoTag)-1])
					log.Fatal(err)
				}
				bufVideoTag = bufVideoTag[:0]
			}
		}
		videoTagCount++
	}
	if len(bufVideoTag) > 0 {
		_, err = db.NamedExec(`insert into video_tags(video_id, tag_id) values(:video_id, :tag_id)`, bufVideoTag)
		if err != nil {
			log.Println(bufVideoTag[0], "...", bufVideoTag[len(bufVideoTag)-1])
			log.Fatal(err)
		}
	}
	log.Println("Inserted all video_tags")
	return nil
}
