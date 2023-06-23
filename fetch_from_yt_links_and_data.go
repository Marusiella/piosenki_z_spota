package main

import (
	"database/sql"
	"errors"
	"flag"
	_ "github.com/go-sql-driver/mysql"
	"sync"
	"time"

	"log"

	"net/http"

	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

type Video struct {
	NameFromSpotify string
	Id              string
	Title           string
	Description     string
	PublishedAt     string
	Thumbnails      []string
}

const developerKey = "HAVE TO SET"

func main() {
	flag.Parse()
	db, err := sql.Open("mysql", "root:password@tcp(192.168.1.123)/songs_spoti")
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(35)
	db.SetMaxIdleConns(10)
	client := &http.Client{
		Transport: &transport.APIKey{Key: developerKey},
	}

	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}

	var wg sync.WaitGroup
	q, er := db.Query("SELECT name_of_song from songs where Title is null")
	if er != nil {
		log.Fatalf("Error query: %v", err)
	}
	var songs []string
	for q.Next() {
		var name string
		err := q.Scan(&name)
		if err != nil {
			log.Fatalf("Error Scan: %v", err)
		}
		songs = append(songs, name)
	}
	for _, x := range songs {
		wg.Add(1)
		x := x
		go func() {
			video, err := getVideo(service, er, x)
			if err != nil {
				log.Fatalf(err.Error())
			}
			addToDbData(db, video)
			log.Printf("Completed %s\n", video.NameFromSpotify)
			wg.Done()
		}()
	}
	wg.Wait()
	defer q.Close()

}

func getVideo(service *youtube.Service, err error, q string) (Video, error) {
	call := service.Search.List(
		[]string{
			"id,snippet",
		}).
		Q(q).
		MaxResults(5)
	response, err := call.Do()
	if err != nil {
		return Video{}, err

	}
	// Group video, channel, and playlist results in separate lists.
	//videos := make(map[string]string)
	videoNameAndId := Video{}
	//L:
	for _, item := range response.Items {
		switch item.Id.Kind {
		case "youtube#video":
			videoNameAndId.Id = item.Id.VideoId
			videoNameAndId.Title = item.Snippet.Title
			videoNameAndId.Description = item.Snippet.Description
			videoNameAndId.PublishedAt = item.Snippet.PublishedAt
			videoNameAndId.NameFromSpotify = q

			var tak []string
			if item.Snippet.Thumbnails.Maxres != nil {
				tak = append(tak, item.Snippet.Thumbnails.Maxres.Url)
			}
			if item.Snippet.Thumbnails.High != nil {
				tak = append(tak, item.Snippet.Thumbnails.High.Url)
			}
			if item.Snippet.Thumbnails.Standard != nil {
				tak = append(tak, item.Snippet.Thumbnails.Standard.Url)
			}
			if item.Snippet.Thumbnails.Default != nil {
				tak = append(tak, item.Snippet.Thumbnails.Default.Url)
			}
			if item.Snippet.Thumbnails.Medium != nil {
				tak = append(tak, item.Snippet.Thumbnails.Medium.Url)
			}

			videoNameAndId.Thumbnails = tak
			return videoNameAndId, nil
			//break L
		}
	}
	return Video{}, errors.New("can't find any video")
}

func addToDbData(conn *sql.DB, video Video) {
	_, err := conn.Exec("UPDATE songs SET Description = ?, PublishedAt = ?, youtubeId = ?, Title = ? where name_of_song = ?", video.Description, video.PublishedAt, video.Id, video.Title, video.NameFromSpotify)
	if err != nil {
		panic(err)
	}
	id := conn.QueryRow("SELECT id from songs where name_of_song = ?", video.NameFromSpotify)
	var idNum int
	err = id.Scan(&idNum)
	if err != nil {
		panic(err)
	}
	for _, x := range video.Thumbnails {
		_, err := conn.Exec("INSERT INTO thumbnails(songId, Thumbnail) values (?,?)", idNum, x)
		if err != nil {
			panic(err)
		}
	}

}
