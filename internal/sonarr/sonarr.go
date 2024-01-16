package sonarr

import (
	"telarr/configuration"

	"github.com/rs/zerolog/log"
	"golift.io/starr"
	"golift.io/starr/sonarr"
)

type Serie struct {
	Title string
	Year  int
}

func GetSeriesList(config configuration.Sonarr) ([]Serie, error) {
	log.Trace().Str("endpoint", config.Endpoint).Msg("contacting radarr")
	c := starr.New(config.ApiKey, config.Endpoint, 0)
	r := sonarr.New(c)

	series, err := r.GetSeries(0)
	if err != nil {
		return nil, err
	}

	// convert the series to the Serie struct
	var seriesList []Serie
	for _, serie := range series {
		seriesList = append(seriesList, Serie{
			Title: serie.Title,
			Year:  serie.Year,
		})
	}

	return seriesList, nil
}
