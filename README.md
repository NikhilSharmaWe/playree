# PLAYREE

**`PLAYREE`** let's you to listen your spotify playlists without ads.

## How it works
1. User logins with their spotify account.
2. Provides Spotify playlist link, which it wants to create.
3. Spotify API fetches list of all tracks and corresponding artists.
4. Youtube API searches for most relevant video and downloads in mp3 and uploads on S3/Minio.
5. User listens to there playlist on Playree without any disturbance.


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

## Demo

https://github.com/NikhilSharmaWe/playree/assets/77074571/49da5bff-1ce2-4e20-b92d-d87caacd45b5





