package cmd

import (
	"log"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/jmoiron/sqlx"
	"github.com/naari3/otomad-tag-sort/pkg/nicovideo"
	"github.com/spf13/cobra"

	_ "github.com/go-sql-driver/mysql"
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

var drop = `
drop table if exists videos;
drop table if exists video_tags;
drop table if exists tags;
`

var schema = `
create table videos (
	video_id varchar(255) not null,
	tags text,
	upload_time datetime not null,
	primary key (video_id)
);
create index videos_upload_time on videos(upload_time);

create table video_tags (
	id int not null auto_increment,
	video_id varchar(255) not null,
	tag_id int not null,
	upload_time datetime,
	primary key (id)
);
create index video_tags_video_id on video_tags(video_id);
create index video_tags_tag_id on video_tags(tag_id);
create index video_tags_upload_time on video_tags(upload_time);

create table tags (
	id int not null auto_increment,
	tag varchar(255) not null unique,
	primary key (id)
);
create index tags_tag on tags(tag);
`

type Video struct {
	VideoID    string    `db:"video_id"`
	Tags       string    `db:"tags"`
	UploadTime time.Time `db:"upload_time"`
}

func (v *Video) NormalizedTags() []string {
	return strings.Split(nicovideo.Normalize(v.Tags), " ")
}

func runFull() error {
	db, err := sqlx.Open("mysql", "root:root@tcp(127.0.0.1:3307)/otomad?parseTime=true&multiStatements=true&collation=utf8mb4_bin")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	collation := "ALTER DATABASE otomad DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;"
	_, err = db.Exec(collation)
	if err != nil {
		log.Fatalf("%q: %s\n", err, collation)
	}

	_, err = db.Exec(drop)
	if err != nil {
		log.Fatalf("%q: %s\n", err, drop)
	}
	_, err = db.Exec(schema)
	if err != nil {
		log.Fatalf("%q: %s\n", err, schema)
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
	uniqueVideoMap := make(map[string]Video)
	for _, v := range videos {
		if _, ok := uniqueVideoMap[v.VideoID]; ok {
			continue
		}
		uniqueVideoMap[v.VideoID] = Video{
			VideoID:    v.VideoID,
			Tags:       v.Tags,
			UploadTime: v.UploadTime,
		}
	}
	log.Println("Collected all unique videos:", len(uniqueVideoMap))
	tagMap := make(map[string]struct{})
	for _, v := range uniqueVideoMap {
		for _, t := range v.NormalizedTags() {
			if _, ok := tagMap[t]; !ok {
				tagMap[t] = struct{}{}
			}
		}
	}
	log.Println("Collected all unique tags:", len(tagMap))

	type Tag struct {
		ID  int    `db:"id"`
		Tag string `db:"tag"`
	}
	tagsBufSize := 50000
	bufTags := make([]Tag, 0, tagsBufSize)
	tagInsertBar := pb.StartNew(len(tagMap))
	for t := range tagMap {
		bufTags = append(bufTags, Tag{Tag: t})
		if len(bufTags) == tagsBufSize {
			_, err = db.NamedExec(`insert into tags(tag) values(:tag)`, bufTags)
			if err != nil {
				log.Fatal(err)
			}
			bufTags = bufTags[:0]
			tagInsertBar.Add(tagsBufSize)
		}
	}
	if len(bufTags) > 0 {
		_, err = db.NamedExec(`insert into tags(tag) values(:tag)`, bufTags)
		if err != nil {
			log.Fatal(err)
		}
		tagInsertBar.Add(len(bufTags))
	}
	tagInsertBar.Finish()
	log.Println("Inserted all tags")
	log.Println("Create tag id map")
	tagMapBar := pb.StartNew(len(tagMap))
	tagIDMap := make(map[string]int)
	rows, err := db.Queryx(`select id, tag from tags`)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var id int
		var tag string
		err = rows.Scan(&id, &tag)
		if err != nil {
			log.Fatal(err)
		}
		tagIDMap[tag] = id
		tagMapBar.Increment()
	}
	tagMapBar.Finish()
	log.Println("Created tag id map")

	videoBufSize := 12500
	bufVideo := make([]Video, 0, videoBufSize)
	videoTagIDsMap := make(map[string][]int)
	videoTagsCount := 0
	videoInsertBar := pb.StartNew(len(uniqueVideoMap))
	for _, v := range uniqueVideoMap {
		bufVideo = append(bufVideo, v)
		if _, ok := videoTagIDsMap[v.VideoID]; !ok {
			videoTagIDsMap[v.VideoID] = make([]int, 0)
		}
		tagIDs := make([]int, 0, len(v.NormalizedTags()))
		for _, t := range v.NormalizedTags() {
			tagIDs = append(tagIDs, tagIDMap[t])
			videoTagsCount++
		}
		videoTagIDsMap[v.VideoID] = tagIDs

		if len(bufVideo) == videoBufSize {
			_, err = db.NamedExec(`insert into videos(video_id, tags, upload_time) values(:video_id, :tags, :upload_time)`, bufVideo)
			if err != nil {
				log.Fatal(err)
			}
			bufVideo = bufVideo[:0]
			videoInsertBar.Add(videoBufSize)
		}
	}
	if len(bufVideo) > 0 {
		_, err = db.NamedExec(`insert into videos(video_id, tags, upload_time) values(:video_id, :tags, :upload_time)`, bufVideo)
		if err != nil {
			log.Fatal(err)
		}
		videoInsertBar.Add(len(bufVideo))
	}
	videoInsertBar.Finish()
	log.Println("Inserted all videos")
	log.Println("Collected all video_tags:", len(videoTagIDsMap))

	type VideoTag struct {
		ID         int       `db:"id"`
		VideoID    string    `db:"video_id"`
		TagID      int       `db:"tag_id"`
		UploadTime time.Time `db:"upload_time"`
	}

	videoTagBufSize := 12500
	bufVideoTag := make([]VideoTag, 0, videoTagBufSize)
	videoTagInsertBar := pb.StartNew(videoTagsCount)
	for videoID, tagIDs := range videoTagIDsMap {
		for _, tagID := range tagIDs {
			bufVideoTag = append(bufVideoTag, VideoTag{
				VideoID:    videoID,
				TagID:      tagID,
				UploadTime: uniqueVideoMap[videoID].UploadTime,
			})
			if len(bufVideoTag) == videoTagBufSize {
				_, err = db.NamedExec(`insert into video_tags(video_id, tag_id) values(:video_id, :tag_id)`, bufVideoTag)
				if err != nil {
					log.Println(bufVideoTag[0], "...", bufVideoTag[len(bufVideoTag)-1])
					log.Fatal(err)
				}
				bufVideoTag = bufVideoTag[:0]
				videoTagInsertBar.Add(videoTagBufSize)
			}
		}
	}
	if len(bufVideoTag) > 0 {
		_, err = db.NamedExec(`insert into video_tags(video_id, tag_id) values(:video_id, :tag_id)`, bufVideoTag)
		if err != nil {
			log.Println(bufVideoTag[0], "...", bufVideoTag[len(bufVideoTag)-1])
			log.Fatal(err)
		}
		videoTagInsertBar.Add(len(bufVideoTag))
	}
	videoTagInsertBar.Finish()
	log.Println("Inserted all video_tags")
	return nil
}
