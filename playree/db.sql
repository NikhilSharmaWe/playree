CREATE TABLE users(
	user_id VARCHAR(50) NOT NULL PRIMARY KEY,
	username VARCHAR(50) NOT NULL
);

CREATE TABLE playlists(
	user_id TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
	playlist_id TEXT NOT NULL,
	playlist_name TEXT NOT NULL,
	PRIMARY KEY (user_id, playlist_id)
);

CREATE TABLE tracks(
	track_key TEXT NOT NULL PRIMARY KEY,
	track_uri TEXT NOT NULL
	playlist_id TEXT NOT NULL REFERENCES playlists(playlist_id) ON DELETE CASCADE,
)
