package sonarr

import (
	"telarr/configuration"
	"telarr/internal/types"

	"github.com/rs/zerolog/log"
	"golift.io/starr"
	"golift.io/starr/sonarr"
)

type Season struct {
	// SeasonNumber is the number of the season.
	SeasonNumber int

	// DownloadedEpisodes is the number of downloaded episodes.
	DownloadedEpisodes int
	// TotalEpisodes is the total number of episodes.
	TotalEpisodes int
}

func GetStatus(config configuration.Sonarr) types.ServiceStatus {
	log.Trace().Str("endpoint", config.Endpoint).Msg("contacting sonarr for status")
	c := starr.New(config.ApiKey, config.Endpoint, 0)
	r := sonarr.New(c)

	status, err := r.GetSystemStatus()
	if err != nil {
		return types.ServiceStatus{
			Name:    "Sonarr",
			Version: "unknown",
			Running: false,
		}
	}

	return types.ServiceStatus{
		Name:    status.AppName,
		Version: status.Version,
		Running: true,
	}
}

// GetSeriesList returns the list of series in the library.
func GetSeriesList(config configuration.Sonarr) ([]Serie, error) {
	log.Trace().Str("endpoint", config.Endpoint).Msg("contacting radarr for series list")
	c := starr.New(config.ApiKey, config.Endpoint, 0)
	r := sonarr.New(c)

	series, err := r.GetSeries(0)
	if err != nil {
		return nil, err
	}

	// convert the series to the Serie struct
	var seriesList []Serie
	for _, serie := range series {
		s := toSerieStruct(serie)
		seriesList = append(seriesList, s)
	}

	return seriesList, nil
}

// GetSerieDetails returns the details of a serie in the library from its name.
func GetSerieDetails(config configuration.Sonarr, serieName string) ([]Serie, error) {
	log.Trace().Str("serieName", serieName).Str("endpoint", config.Endpoint).Msg("contacting radarr for serie details")
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
				s := libSerie
				seriesList = append(seriesList, s)
			}
		}
	}

	return seriesList, nil
}

// GetSerieName returns the name of a serie in the library from its id.
func GetSerieName(config configuration.Sonarr, serieId int) (string, error) {
	log.Trace().Int("serieId", serieId).Str("endpoint", config.Endpoint).Msg("contacting radarr for serie name")
	c := starr.New(config.ApiKey, config.Endpoint, 0)
	r := sonarr.New(c)

	serie, err := r.GetSeriesByID(int64(serieId))
	if err != nil {
		return "", err
	}

	return serie.Title, nil
}

// RemoveSerie removes a serie from the library.
func RemoveSerie(config configuration.Sonarr, serieId int) error {
	log.Trace().Int("serieId", serieId).Str("endpoint", config.Endpoint).Msg("contacting radarr to remove serie")
	c := starr.New(config.ApiKey, config.Endpoint, 0)
	r := sonarr.New(c)

	err := r.DeleteSeries(serieId, true, false)
	if err != nil {
		return err
	}

	return nil
}

func LookupSerie(config configuration.Sonarr, serieName string) ([]Serie, error) {
	log.Trace().Str("serieName", serieName).Str("endpoint", config.Endpoint).Msg("contacting radarr for serie details")
	c := starr.New(config.ApiKey, config.Endpoint, 0)
	r := sonarr.New(c)

	series, err := r.Lookup(serieName)
	if err != nil {
		return nil, err
	}

	// convert the series to the Serie struct
	var seriesList []Serie
	for _, serie := range series {
		s := toSerieStruct(serie)
		seriesList = append(seriesList, s)
	}

	return seriesList, nil
}

func AddSerie(config configuration.Sonarr, serie Serie, qualityProfileId int64) error {
	log.Trace().Str("serieTitle", serie.Title).Str("endpoint", config.Endpoint).Msg("contacting radarr to add serie")
	c := starr.New(config.ApiKey, config.Endpoint, 0)
	r := sonarr.New(c)

	_, err := r.AddSeries(&sonarr.AddSeriesInput{
		Title:            serie.Title,
		TvdbID:           serie.TvdbId,
		QualityProfileID: qualityProfileId,
		Monitored:        true,
		RootFolderPath:   "/tv",
		AddOptions: &sonarr.AddSeriesOptions{
			SearchForMissingEpisodes: true,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func GetQualityProfiles(config configuration.Sonarr) ([]types.QualityProfile, error) {
	log.Trace().Str("endpoint", config.Endpoint).Msg("contacting radarr for quality profiles")
	c := starr.New(config.ApiKey, config.Endpoint, 0)
	s := sonarr.New(c)

	profiles, err := s.GetQualityProfiles()
	if err != nil {
		return nil, err
	}

	// convert the profiles to the QualityProfile struct
	var profilesList []types.QualityProfile
	for _, profile := range profiles {
		p := types.QualityProfile{
			ID:   profile.ID,
			Name: profile.Name,
		}
		profilesList = append(profilesList, p)
	}

	return profilesList, nil
}

func GetQualityProfileId(config configuration.Sonarr, profileName string) (int64, error) {
	log.Trace().Str("profileName", profileName).Str("endpoint", config.Endpoint).Msg("contacting radarr for quality profile id")
	c := starr.New(config.ApiKey, config.Endpoint, 0)
	s := sonarr.New(c)

	profiles, err := s.GetQualityProfiles()
	if err != nil {
		return -1, err
	}

	for _, profile := range profiles {
		if profile.Name == profileName {
			return profile.ID, nil
		}
	}

	return -1, nil
}

/* Tools */

func toSerieStruct(serie *sonarr.Series) Serie {
	s := Serie{
		TvdbId:            serie.TvdbID,
		SerieId:           serie.ID,
		IsInLibrary:       serie.ID > 0,
		Title:             serie.Title,
		Year:              serie.Year,
		TotalSeasonsCount: serie.Statistics.SeasonCount,
		Rating:            serie.Ratings.Value,
		NumberOfVotes:     int(serie.Ratings.Votes),
		Overview:          serie.Overview,
		Genres:            serie.Genres,
		Downloaded:        serie.Statistics.EpisodeFileCount > 0,
		Size:              float64(serie.Statistics.SizeOnDisk) / 1024 / 1024 / 1024,
	}

	// get the seasons
	for _, season := range serie.Seasons {
		se := Season{
			SeasonNumber: season.SeasonNumber,
		}
		if season.Statistics != nil {
			se.DownloadedEpisodes = season.Statistics.EpisodeFileCount
			se.TotalEpisodes = season.Statistics.TotalEpisodeCount
		}

		s.Seasons = append(s.Seasons, se)
	}

	// get the cover image
	for _, image := range serie.Images {
		if image.CoverType == "poster" {
			s.CoverImage = image.RemoteURL
			break
		}
	}

	return s
}
