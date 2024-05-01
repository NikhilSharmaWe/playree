CREATE TABLE users(
	user_id VARCHAR(50) NOT NULL PRIMARY KEY,
	username VARCHAR(50) NOT NULL
);

CREATE TABLE playlists(
	playlist_id TEXT NOT NULL PRIMARY KEY,
	playlist_name TEXT NOT NULL ,
	user_id TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE TABLE tracks (
	track_key TEXT NOT NULL PRIMARY KEY,
  	track_uri TEXT NOT NULL,
  	playlist_id TEXT NOT NULL REFERENCES playlists(playlist_id) ON DELETE CASCADE,
  	inserted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tracks_on_playlist_id ON tracks(playlist_id);
