package main

import (
	"database/sql"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"sync"
	"time"
)

func main() {
	db, err := sql.Open("mysql", "root:password@tcp(192.168.1.123)/songs_spoti")
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(35)
	db.SetMaxIdleConns(10)
	file, err := os.ReadFile("message.json")
	if err != nil {
		panic(err)
	}
	//text := string(file)
	//fmt.Println(text)
	var xx []string
	err = json.Unmarshal(file, &xx)
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	for _, x := range xx {
		x := x
		wg.Add(1)
		go func() {
			_, err := db.Exec("INSERT INTO songs(name_of_song) values (?);", x)
			if err != nil {
				panic("Can't insert " + err.Error())
			}
			wg.Done()
		}()
	}
	wg.Wait()
	err = db.Close()
	if err != nil {
		panic("Can't close connections!!!")
	}
}
