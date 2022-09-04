package cmd

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/naari3/otomad-tag-sort/pkg/nicovideo"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update database",
	Long:  `Update database`,
	Run: func(cmd *cobra.Command, args []string) {
		runUpdate()
	},
}

func runUpdate() error {
	db, err := sqlx.Open("sqlite3", "./niconico.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	videos, err := nicovideo.ReadAllVideoFromDirectory("jsonl")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Collected all videos:", len(videos))

	append_videos, err := nicovideo.ReadAllVideoFromDirectory("jsonl_append")
	if err != nil {
		panic(err)
	}
	log.Println("Collected all append_videos:", len(append_videos))

	videos = append(videos, append_videos...)

	var videoIDsFromDB []string
	err = db.Select(&videoIDsFromDB, "select video_id from videos")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Collected all videos from db:", len(videoIDsFromDB))

	videoIDMap := make(map[string]bool)
	for _, videoID := range videoIDsFromDB {
		videoIDMap[videoID] = true
	}

	newVideos := make([]nicovideo.Video, 0)
	for _, video := range videos {
		if _, ok := videoIDMap[video.VideoID]; !ok {
			newVideos = append(newVideos, video)
		}
	}

	log.Println("New videos:", len(newVideos))

	for _, video := range newVideos {
		_, err = db.NamedExec("insert into videos (video_id, tags, upload_time) values (:video_id, :tags, :upload_time)", video)
		if err != nil {
			log.Fatal(err)
		}
	}

	tagMap := make(map[string]bool)
	for _, video := range newVideos {
		for _, tag := range video.NormalizedTags() {
			if _, ok := tagMap[tag]; !ok {
				tagMap[tag] = true
			}
		}
	}

	tagIDMap := make(map[string]int)
	for tag := range tagMap {
		tag = nicovideo.Normalize(tag)
		// check if tag exists
		var count int
		err = db.Get(&count, "select count(*) from tags where tag = ?", tag)
		if err != nil {
			log.Fatal(err)
		}
		if count == 0 {
			result, err := db.Exec("insert into tags (tag) values (?)", tag)
			if err != nil {
				log.Fatal(err)
			}
			newID, err := result.LastInsertId()
			if err != nil {
				log.Fatal(err)
			}
			tagIDMap[tag] = int(newID)
		}
	}

	for _, video := range newVideos {
		for _, tag := range video.NormalizedTags() {
			tag = nicovideo.Normalize(tag)
			_, err = db.Exec("insert into video_tags (video_id, tag_id) values (?, ?)", video.VideoID, tagIDMap[tag])
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	return nil
}
