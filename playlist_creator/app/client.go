package app

import (
	"github.com/NikhilSharmaWe/playree/playlist_creator/proto"
	"google.golang.org/grpc"
)

func NewCreatePlaylistClient(remoteAddr string) (proto.CreatePlaylistServiceClient, error) {
	conn, err := grpc.Dial(remoteAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	c := proto.NewCreatePlaylistServiceClient(conn)
	return c, nil
}
