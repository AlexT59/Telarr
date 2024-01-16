package radarr

import (
	"telarr/configuration"

	"github.com/rs/zerolog/log"
	"golift.io/starr"
	"golift.io/starr/radarr"
)

type Film struct {
	Title string
	Year  int
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
			Title: film.Title,
			Year:  film.Year,
		})
	}

	return filmsList, nil
}
