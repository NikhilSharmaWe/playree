package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/NikhilSharmaWe/playree/playlist_creator/app"
	"github.com/NikhilSharmaWe/playree/playlist_creator/proto"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load("vars.env"); err != nil {
		log.Fatal(err)
	}
}

func main() {
	application, err := app.NewApplication()
	if err != nil {
		log.Fatal(err)
	}

	tracks := []*proto.Track{
		{
			Name:    "Blackholesun",
			Artists: "Soundgarden",
		},
		{
			Name:    "Like a Stone",
			Artists: "Audioslave",
		},
	}

	req := proto.CreatePlaylistRequest{
		PlayreePlaylistId: "1234",
		Tracks:            tracks,
	}

	go func() {
		time.Sleep(3 * time.Second)
		resp, err := application.CreatePlaylistClient.CreatePlaylist(context.Background(), &req)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("RESPONSE: ", resp)
	}()

	log.Fatal(application.MakeCreatePlaylistServerAndRun())
}
