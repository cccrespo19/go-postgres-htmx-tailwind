package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

type Anime struct {
	Id       int
	Name     string
	Author   string
	Episodes int
}

func main() {
	// regexp replace to replace all special characters with - for Anime.Path
	// sample := "Jojo's Bizzare Adventure"
	// re := regexp.MustCompile(`[^a-zA-Z]`)
	// s := re.ReplaceAllString(sample, "-")
	// fmt.Println(s)

	// godotenv setup for db url
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	connStr := os.Getenv("DB_URL")
	// open connection for db
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	// defer close connection
	defer db.Close()

	// create table in db for anime collection
	// safe to run function on every load sql query has CREATE TABLE IF NOT EXISTS
	createAnimeTable(db)

	// handle default path
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			animes := getAnimes(db)

			tmpl := template.Must(template.ParseFiles("index.html"))
			tmpl.Execute(w, animes)
		}
	})

	http.HandleFunc("/animes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			name := r.PostFormValue("name")
			author := r.PostFormValue("author")
			episodes, err := strconv.Atoi(r.PostFormValue("episodes"))
			if err != nil {
				log.Fatal(err)
			}
			newAnime := Anime{0, name, author, episodes}
			insertAnime(db, newAnime)

			animes := getAnimes(db)

			tmpl := template.Must(template.ParseFiles("index.html"))
			tmpl.ExecuteTemplate(w, "content", animes)
		}
	})

	http.HandleFunc("/animes/", func(w http.ResponseWriter, r *http.Request) {
		animeId, err := strconv.Atoi(strings.Replace(r.URL.Path, "/animes/", "", 1))
		if err != nil {
			log.Fatal(nil)
		}

		if r.Method == "DELETE" {
			deleteAnime(db, animeId)

			animes := getAnimes(db)

			tmpl := template.Must(template.ParseFiles("index.html"))
			tmpl.ExecuteTemplate(w, "content", animes)
		}

		if r.Method == "PUT" {
			name := r.PostFormValue("name")
			author := r.PostFormValue("author")
			episodes, err := strconv.Atoi(r.PostFormValue("episodes"))
			if err != nil {
				log.Fatal(err)
			}
			newValues := Anime{0, name, author, episodes}
			updateAnime(db, animeId, newValues)

			animes := getAnimes(db)

			tmpl := template.Must(template.ParseFiles("index.html"))
			tmpl.ExecuteTemplate(w, "content", animes)
		}
	})

	http.HandleFunc("/update-form/", func(w http.ResponseWriter, r *http.Request) {
		animeId, err := strconv.Atoi(strings.Replace(r.URL.Path, "/update-form/", "", 1))
		if err != nil {
			log.Fatal(err)
		}

		anime := getAnime(db, animeId)

		tmpl := template.Must(template.ParseFiles("update-form.html"))
		tmpl.Execute(w, anime)
	})

	log.Fatal(http.ListenAndServe(":3000", nil))
}

func createAnimeTable(db *sql.DB) {
	query := `CREATE TABLE IF NOT EXISTS anime(
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		author VARCHAR(100) NOT NULL,
		episodes INTEGER NOT NULL,
		created timestamp DEFAULT NOW()
	)`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func insertAnime(db *sql.DB, anime Anime) int {
	query := `INSERT INTO anime (name, author, episodes)
		VALUES ($1, $2, $3) RETURNING id`
	var animeId int
	err := db.QueryRow(query, anime.Name, anime.Author, anime.Episodes).Scan(&animeId)
	if err != nil {
		log.Fatal(err)
	}
	return animeId
}

func updateAnime(db *sql.DB, animeId int, newValues Anime) {
	query := `UPDATE anime
		SET name = $1, author = $2, episodes = $3
		WHERE id = $4`
	_, err := db.Exec(query, newValues.Name, newValues.Author, newValues.Episodes, animeId)
	if err != nil {
		log.Fatal(err)
	}
}

func deleteAnime(db *sql.DB, animeId int) {
	query := `DELETE FROM anime
		WHERE id = $1`
	_, err := db.Exec(query, animeId)
	if err != nil {
		log.Fatal(err)
	}
}

func getAnimes(db *sql.DB) []Anime {
	animes := []Anime{}
	query := `SELECT id, name, author, episodes FROM anime
		ORDER BY created`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var id int
	var name string
	var author string
	var episodes int
	for rows.Next() {
		err := rows.Scan(&id, &name, &author, &episodes)
		if err != nil {
			log.Fatal(err)
		}
		animes = append(animes, Anime{id, name, author, episodes})
	}
	return animes
}

func getAnime(db *sql.DB, animeId int) Anime {
	var name string
	var author string
	var episodes int
	query := `SELECT name, author, episodes FROM anime
		WHERE id = $1`
	err := db.QueryRow(query, animeId).Scan(&name, &author, &episodes)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Fatalf("No rows found with ID %d", animeId)
		}
		log.Fatal(err)
	}
	anime := Anime{animeId, name, author, episodes}
	return anime
}
