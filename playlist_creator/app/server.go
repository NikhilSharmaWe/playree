package app

import (
	"context"
	"net"

	"github.com/NikhilSharmaWe/playree/playlist_creator/proto"
	"google.golang.org/grpc"
)

func (app *Application) MakeCreatePlaylistServerAndRun() error {
	createPlaylistServer := NewCreatePlaylistServer(app)
	ln, err := net.Listen("tcp", app.Addr)
	if err != nil {
		return err
	}

	opts := []grpc.ServerOption{}
	server := grpc.NewServer(opts...)
	proto.RegisterCreatePlaylistServiceServer(server, createPlaylistServer)

	return server.Serve(ln)
}

type CreatePlaylistServer struct {
	svc CreatePlaylistService
	proto.UnimplementedCreatePlaylistServiceServer
}

func NewCreatePlaylistServer(app *Application) *CreatePlaylistServer {
	return &CreatePlaylistServer{
		svc: NewCreatePlaylistService(app),
	}
}

func (s *CreatePlaylistServer) CreatePlaylist(ctx context.Context, req *proto.CreatePlaylistRequest) (*proto.CreatePlaylistResponse, error) {
	if err := s.svc.CreatePlaylist(CreatePlaylistRequest{
		PlayreePlaylistID: req.PlayreePlaylistId,
		Tracks:            req.Tracks,
	}); err != nil {
		return nil, err
	}

	return &proto.CreatePlaylistResponse{
		PlayreePlaylistId: req.PlayreePlaylistId,
	}, nil
}
