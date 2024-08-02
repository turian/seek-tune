package utils

import (
	"database/sql"
	"errors"
	"fmt"
	"song-recognition/models"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// DbClient represents an SQLite client
type DbClient struct {
	db *sql.DB
}

// NewDbClient creates a new instance of DbClient
func NewDbClient() (*DbClient, error) {
	db, err := sql.Open("sqlite3", "./song_recognition.db")
	if err != nil {
		return nil, fmt.Errorf("error connecting to SQLite: %v", err)
	}

	return &DbClient{db: db}, nil
}

// Close closes the underlying SQLite connection
func (db *DbClient) Close() error {
	if db.db != nil {
		return db.db.Close()
	}
	return nil
}

func (db *DbClient) StoreFingerprints(fingerprints map[uint32]models.Couple) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}

	for address, couple := range fingerprints {
		_, err := tx.Exec(`
			INSERT INTO fingerprints (address, anchorTimeMs, songID)
			VALUES (?, ?, ?)
			ON CONFLICT(address) DO UPDATE SET anchorTimeMs = excluded.anchorTimeMs, songID = excluded.songID`,
			address, couple.AnchorTimeMs, couple.SongID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error inserting/updating fingerprint: %v", err)
		}
	}

	return tx.Commit()
}

func (db *DbClient) GetCouples(addresses []uint32) (map[uint32][]models.Couple, error) {
	couples := make(map[uint32][]models.Couple)

	for _, address := range addresses {
		rows, err := db.db.Query(`
			SELECT anchorTimeMs, songID FROM fingerprints WHERE address = ?`, address)
		if err != nil {
			return nil, fmt.Errorf("error querying fingerprints: %v", err)
		}
		defer rows.Close()

		var docCouples []models.Couple
		for rows.Next() {
			var couple models.Couple
			if err := rows.Scan(&couple.AnchorTimeMs, &couple.SongID); err != nil {
				return nil, fmt.Errorf("error scanning couple: %v", err)
			}
			docCouples = append(docCouples, couple)
		}

		couples[address] = docCouples
	}

	return couples, nil
}

func (db *DbClient) TotalSongs() (int, error) {
	var count int
	err := db.db.QueryRow(`SELECT COUNT(*) FROM songs`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting songs: %v", err)
	}
	return count, nil
}

func (db *DbClient) RegisterSong(songTitle, songArtist, ytID string) (uint32, error) {
	songID := GenerateUniqueID()
	key := GenerateSongKey(songTitle, songArtist)

	_, err := db.db.Exec(`
		INSERT INTO songs (id, title, artist, ytID, key)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(key) DO NOTHING`,
		songID, songTitle, songArtist, ytID, key)
	if err != nil {
		return 0, fmt.Errorf("failed to register song: %v", err)
	}

	return songID, nil
}

type Song struct {
	Title     string
	Artist    string
	YouTubeID string
}

const FILTER_KEYS = "id | ytID | key"

func (db *DbClient) GetSong(filterKey string, value interface{}) (s Song, songExists bool, e error) {
	if !strings.Contains(FILTER_KEYS, filterKey) {
		return Song{}, false, errors.New("invalid filter key")
	}

	query := fmt.Sprintf("SELECT title, artist, ytID FROM songs WHERE %s = ?", filterKey)
	row := db.db.QueryRow(query, value)

	var song Song
	err := row.Scan(&song.Title, &song.Artist, &song.YouTubeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return Song{}, false, nil
		}
		return Song{}, false, fmt.Errorf("failed to retrieve song: %v", err)
	}

	return song, true, nil
}

func (db *DbClient) GetSongByID(songID uint32) (Song, bool, error) {
	return db.GetSong("id", songID)
}

func (db *DbClient) GetSongByYTID(ytID string) (Song, bool, error) {
	return db.GetSong("ytID", ytID)
}

func (db *DbClient) GetSongByKey(key string) (Song, bool, error) {
	return db.GetSong("key", key)
}

func (db *DbClient) DeleteSongByID(songID uint32) error {
	_, err := db.db.Exec(`DELETE FROM songs WHERE id = ?`, songID)
	if err != nil {
		return fmt.Errorf("failed to delete song: %v", err)
	}
	return nil
}

func (db *DbClient) DeleteCollection(tableName string) error {
	_, err := db.db.Exec(fmt.Sprintf(`DELETE FROM %s`, tableName))
	if err != nil {
		return fmt.Errorf("error deleting table: %v", err)
	}
	return nil
}
