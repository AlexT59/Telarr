package sonarr

import (
	"strconv"
	"telarr/configuration"

	"github.com/rs/zerolog/log"
	"golift.io/starr"
	"golift.io/starr/sonarr"
)

type Serie struct {
	TvdbId  int64
	SerieId int64

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

	// Rating is the rating of the serie (/10).
	Rating float64
	// NumberOfVotes is the number of votes for the serie.
	NumberOfVotes int

	// Overview is the overview of the serie.
	Overview string
	// Genres is the list of genres of the serie.
	Genres []string
	// Studio is the studio of the serie.
	Studio string

	// Size is the size of the serie on disk GB.
	Size float64
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

/* Serie */

func (s Serie) PrintSerieTitle() string {
	str := "📺 *" + s.Title + "* (_" + strconv.Itoa(s.Year) + "_)"

	return str
}

func (s Serie) PrintSerieDetails() string {
	str := "📅 " + "*Release year*: " + strconv.Itoa(s.Year) + "\n"
	str += "⭐ " + "*Rating*: " + strconv.FormatFloat(s.Rating, 'f', 1, 64) + "/10 (_" + strconv.Itoa(s.NumberOfVotes) + " votes_)\n"
	str += "🎞 " + "*Genres*: "
	for i, genre := range s.Genres {
		if i != 0 {
			str += ", "
		}
		str += genre
	}
	str += "\n"
	str += "📝 " + "*Overview*: " + s.Overview + "\n"
	str += "\n"

	str += "💾 " + "*Size*: " + strconv.FormatFloat(s.Size, 'f', 2, 64) + " GB\n"

	str += "*Seasons* (" + strconv.Itoa(s.TotalSeasonsCount) + "): \n"
	for _, season := range s.Seasons {
		if season.SeasonNumber == 0 {
			str += "\t- _Specials "
		} else {
			str += "\t- _Season " + strconv.Itoa(season.SeasonNumber)
		}
		str += " (" + strconv.Itoa(season.DownloadedEpisodes) + "/" + strconv.Itoa(season.TotalEpisodes) + ")_\n"
	}

	return str
}

/* Tools */

func toSerieStruct(serie *sonarr.Series) Serie {
	s := Serie{
		TvdbId:            serie.TvdbID,
		SerieId:           serie.ID,
		Title:             serie.Title,
		Year:              serie.Year,
		TotalSeasonsCount: serie.Statistics.SeasonCount,
		Rating:            serie.Ratings.Value,
		NumberOfVotes:     int(serie.Ratings.Votes),
		Overview:          serie.Overview,
		Genres:            serie.Genres,
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
