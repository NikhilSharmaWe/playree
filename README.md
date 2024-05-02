# PLAYREE

**`PLAYREE`** let's you to listen your spotify playlists without ads.

## Tools
- **Backend Languages**: Golang
- **Frontend Languages**: JavaScript
- **Database**: PostgreSQL, Redis
- **Message Broker**: RabbitMQ
- **Authentication**: Spotify OAuth
- **Session Management**: Cookies and Sessions
- **Microservices Communication**: gRPC
- **Object Storage**: MinIO (S3-compatible)


## Services
1. **Playree** (web server)

   Playree serves as the web server that `users interact with`. It is responsible for handling user requests to create new playlists from there spotify playlists or listening playlists.
   Playree communicates with playlist-creator `gRPC` server for handling the create playlist requests.
   Requests and responses are sent through `RabbitMQ` between both the servers.

2. **Playlist-Creator** (gRPC server)

   The Playlist-Creator is a `gRPC` server which handles the create playlist request.
   The Request contains the list of track names and corresponding artists in the playlist.
   Service first fetches the top `Youtube` video relevant with the song name and artist, downloads it in mp3 and uploads it to S3/Minio in a separate folder and sends the response to Playree about the status.


