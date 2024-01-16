package sonarr

import (
	"telarr/configuration"

	"github.com/rs/zerolog/log"
	"golift.io/starr"
	"golift.io/starr/sonarr"
)

type Serie struct {
	// Title is the title of the serie.
	Title string
	// Year is the release year of the serie.
	Year int

	// Seasons is the list of seasons of the serie.
	Seasons []Season
	// TotalSeasonsCount is the total number of seasons.
	TotalSeasonsCount int

	// CoverImage is the url to the cover of the serie.
	CoverImage string
}

type Season struct {
	// SeasonNumber is the number of the season.
	SeasonNumber int

	// DownloadedEpisodes is the number of downloaded episodes.
	DownloadedEpisodes int
	// TotalEpisodes is the total number of episodes.
	TotalEpisodes int
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
		s := Serie{
			Title:             serie.Title,
			Year:              serie.Year,
			TotalSeasonsCount: serie.Statistics.SeasonCount,
		}

		// get the seasons
		for _, season := range serie.Seasons {
			s.Seasons = append(s.Seasons, Season{
				SeasonNumber:       season.SeasonNumber,
				DownloadedEpisodes: season.Statistics.EpisodeFileCount,
				TotalEpisodes:      season.Statistics.TotalEpisodeCount,
			})
		}

		seriesList = append(seriesList, s)
	}

	return seriesList, nil
}

func GetSerieDetails(config configuration.Sonarr, serieName string) ([]Serie, error) {
	log.Trace().Str("endpoint", config.Endpoint).Msg("contacting radarr")
	c := starr.New(config.ApiKey, config.Endpoint, 0)
	r := sonarr.New(c)

	series, err := r.Lookup(serieName)
	if err != nil {
		return nil, err
	}

	// convert the series to the Serie struct
	var seriesList []Serie
	libSeries, err := GetSeriesList(config)
	if err != nil {
		return nil, err
	}
	for _, serie := range series {
		for _, libSerie := range libSeries {
			// keep only series that are in the library
			if serie.Title == libSerie.Title {
				s := Serie{
					Title:             serie.Title,
					Year:              serie.Year,
					TotalSeasonsCount: serie.Statistics.SeasonCount,
					Seasons:           libSerie.Seasons,
				}

				// get the cover image
				for _, image := range serie.Images {
					if image.CoverType == "poster" {
						s.CoverImage = image.RemoteURL
						break
					}
				}

				seriesList = append(seriesList, s)
			}
		}
	}

	return seriesList, nil
}
