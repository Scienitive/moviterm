package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func (app *App) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Decoding JSON into Movie struct
	var movie Movie
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Setting up a transaction for multiple statements
	tx, err := app.db.Begin()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	// Parse the URL to get the ID
	idStr := r.URL.Path[len("/update/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		tx.Rollback()
		return
	}

	// Update the movie
	_, err = tx.Exec(
		`UPDATE movies
		SET title = ?, year = ?, rating = ?, imdbRating = ?
		WHERE id = ?`,
		movie.Title, movie.Year, movie.Rating, movie.ImdbRating, id,
	)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println(err.Error())
		tx.Rollback()
		return
	}

	// Delete genres
	_, err = tx.Exec(
		`DELETE FROM movies_genres
		WHERE movieId = ?`,
		id,
	)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println(err.Error())
		tx.Rollback()
		return
	}

	// Recreate genres
	for _, genre := range movie.Genres {
		_, err := tx.Exec(
			`INSERT OR IGNORE INTO genres (name)
			VALUES (?)`,
			genre,
		)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Println(err.Error())
			tx.Rollback()
			return
		}
		var genreId int
		row := tx.QueryRow(
			`SELECT id FROM genres
			WHERE name = ?`,
			genre,
		)
		err = row.Scan(&genreId)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Println(err.Error())
			tx.Rollback()
			return
		}
		_, err = tx.Exec(
			`INSERT OR IGNORE INTO movies_genres (movieId, genreId)
			VALUES (?, ?)`,
			id, genreId,
		)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Println(err.Error())
			tx.Rollback()
			return
		}
	}

	// Delete directors
	_, err = tx.Exec(
		`DELETE FROM movies_directors
		WHERE movieId = ?`,
		id,
	)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println(err.Error())
		tx.Rollback()
		return
	}

	// Recreate directors
	for _, director := range movie.Directors {
		_, err := tx.Exec(
			`INSERT OR IGNORE INTO directors (name)
			VALUES (?)`,
			director,
		)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Println(err.Error())
			tx.Rollback()
			return
		}
		var directorId int
		row := tx.QueryRow(
			`SELECT id FROM directors
			WHERE name = ?`,
			director,
		)
		err = row.Scan(&directorId)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Println(err.Error())
			tx.Rollback()
			return
		}
		_, err = tx.Exec(
			`INSERT OR IGNORE INTO movies_directors (movieId, genreId)
			VALUES (?, ?)`,
			id, directorId,
		)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Println(err.Error())
			tx.Rollback()
			return
		}
	}

	// Commit all changes
	if err := tx.Commit(); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	fmt.Fprintln(w, "Movie successfully updated.")
}