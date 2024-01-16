package radarr

import (
	"telarr/configuration"

	"github.com/rs/zerolog/log"
	"golift.io/starr"
	"golift.io/starr/radarr"
)

type Film struct {
	TmdbId int64
	// Title is the title of the film.
	Title string
	// Year is the release year of the film.
	Year int

	// CoverImage is the url to the cover of the film.
	CoverImage string
}

func GetFilmsList(config configuration.Radarr) ([]Film, error) {
	log.Trace().Str("endpoint", config.Endpoint).Msg("contacting radarr")
	c := starr.New(config.ApiKey, config.Endpoint, 0)
	r := radarr.New(c)

	films, err := r.GetMovie(0)
	if err != nil {
		return nil, err
	}

	// convert the films to the Film struct
	var filmsList []Film
	for _, film := range films {
		filmsList = append(filmsList, Film{
			TmdbId: film.TmdbID,
			Title:  film.Title,
			Year:   film.Year,
		})
	}

	return filmsList, nil
}

func GetFilmDetails(config configuration.Radarr, movieName string) ([]Film, error) {
	log.Trace().Str("endpoint", config.Endpoint).Msg("contacting radarr")
	c := starr.New(config.ApiKey, config.Endpoint, 0)
	r := radarr.New(c)

	films, err := r.Lookup(movieName)
	if err != nil {
		return nil, err
	}

	// convert the films to the Film struct
	var filmsList []Film
	libFilms, err := GetFilmsList(config)
	if err != nil {
		return nil, err
	}
	for _, film := range films {
		for _, libFilm := range libFilms {
			// keep only films that are in the library
			if film.TmdbID == libFilm.TmdbId {
				f := Film{
					Title: film.Title,
					Year:  film.Year,
				}
				for _, image := range film.Images {
					if image.CoverType == "poster" {
						f.CoverImage = image.RemoteURL
						break
					}
				}

				filmsList = append(filmsList, f)
			}
		}
	}

	return filmsList, nil
}
