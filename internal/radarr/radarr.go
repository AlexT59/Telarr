package radarr

import (
	"telarr/configuration"
	"telarr/internal/types"

	"github.com/rs/zerolog/log"
	"golift.io/starr"
	"golift.io/starr/radarr"
)

func GetStatus(config configuration.Radarr) types.ServiceStatus {
	log.Trace().Str("endpoint", config.Endpoint).Msg("contacting radarr for status")
	c := starr.New(config.ApiKey, config.Endpoint, 0)
	r := radarr.New(c)

	status, err := r.GetSystemStatus()
	if err != nil {
		return types.ServiceStatus{
			Name:    "Radarr",
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

// GetFilmsList returns the list of films in the library.
func GetFilmsList(config configuration.Radarr) ([]Film, error) {
	log.Trace().Str("endpoint", config.Endpoint).Msg("contacting radarr for movie list")
	c := starr.New(config.ApiKey, config.Endpoint, 0)
	r := radarr.New(c)

	films, err := r.GetMovie(0)
	if err != nil {
		return nil, err
	}

	// convert the films to the Film struct
	var filmsList []Film
	for _, film := range films {
		filmsList = append(filmsList, toFilmStruct(film))
	}

	return filmsList, nil
}

// GetFilmDetails returns the details of a film in the library from its name.
func GetFilmDetails(config configuration.Radarr, movieName string) ([]Film, error) {
	log.Trace().Str("movieName", movieName).Str("endpoint", config.Endpoint).Msg("contacting radarr for movie details")
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
				f := toFilmStruct(film)
				filmsList = append(filmsList, f)
			}
		}
	}

	return filmsList, nil
}

// GetMovieName returns the name of a movie in the library from its id.
func GetMovieName(config configuration.Radarr, movieId int) (string, error) {
	log.Trace().Int("movieId", movieId).Str("endpoint", config.Endpoint).Msg("contacting radarr for movie name")
	c := starr.New(config.ApiKey, config.Endpoint, 0)
	r := radarr.New(c)

	movie, err := r.GetMovieByID(int64(movieId))
	if err != nil {
		return "", err
	}

	return movie.Title, nil
}

// RemoveFilm removes a film from the library.
func RemoveFilm(config configuration.Radarr, movieId int) error {
	log.Trace().Int("movieId", movieId).Str("endpoint", config.Endpoint).Msg("contacting radarr	to remove movie")
	c := starr.New(config.ApiKey, config.Endpoint, 0)
	r := radarr.New(c)

	err := r.DeleteMovie(int64(movieId), true, false)
	if err != nil {
		return err
	}

	return nil
}

// LookupFilm looks for a film in radarr.
func LookupFilm(config configuration.Radarr, movieName string) ([]Film, error) {
	log.Trace().Str("movieName", movieName).Str("endpoint", config.Endpoint).Msg("contacting radarr for movie lookup")
	c := starr.New(config.ApiKey, config.Endpoint, 0)
	r := radarr.New(c)

	films, err := r.Lookup(movieName)
	if err != nil {
		return nil, err
	}

	// convert the films to the Film struct
	var filmsList []Film
	for _, film := range films {
		filmsList = append(filmsList, toFilmStruct(film))
	}

	return filmsList, nil
}

func AddFilm(config configuration.Radarr, film Film, qualityProfileId int64) (int64, error) {
	log.Trace().Str("movieName", film.Title).Str("endpoint", config.Endpoint).Msg("contacting radarr to add movie")
	c := starr.New(config.ApiKey, config.Endpoint, 0)
	r := radarr.New(c)

	newFilm, err := r.AddMovie(&radarr.AddMovieInput{
		Title:            film.Title,
		TmdbID:           film.TmdbId,
		QualityProfileID: qualityProfileId,
		Monitored:        true,
		RootFolderPath:   "/movies",
		AddOptions: &radarr.AddMovieOptions{
			SearchForMovie: true,
			Monitor:        "movieOnly",
		},
	})
	if err != nil {
		return -1, err
	}

	return newFilm.ID, nil
}

func GetQualityProfiles(config configuration.Radarr) ([]types.QualityProfile, error) {
	log.Trace().Str("endpoint", config.Endpoint).Msg("contacting radarr for quality profiles")
	c := starr.New(config.ApiKey, config.Endpoint, 0)
	r := radarr.New(c)

	profiles, err := r.GetQualityProfiles()
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

func GetQualityProfileId(config configuration.Radarr, profileName string) (int64, error) {
	log.Trace().Str("profileName", profileName).Str("endpoint", config.Endpoint).Msg("contacting radarr for quality profile id")
	c := starr.New(config.ApiKey, config.Endpoint, 0)
	r := radarr.New(c)

	profiles, err := r.GetQualityProfiles()
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

// GetDownloadingStatus returns the downloading status of a film.
func GetDownloadingStatus(config configuration.Radarr, filmId int) (types.DownloadingStatus, error) {
	log.Trace().Str("endpoint", config.Endpoint).Msg("contacting radarr for downloading status")
	c := starr.New(config.ApiKey, config.Endpoint, 0)
	r := radarr.New(c)

	_, err := r.SendCommand(&radarr.CommandRequest{
		Name: "RefreshMonitoredDownloads",
	})
	if err != nil {
		return types.DownloadingStatus{}, err
	}
	queue, err := r.GetQueue(0, 100)
	if err != nil {
		return types.DownloadingStatus{}, err
	}

	// get our film in the queue
	for _, rec := range queue.Records {
		if rec.MovieID == int64(filmId) {
			return types.DownloadingStatus{
				Found:                   true,
				FilmId:                  rec.MovieID,
				Status:                  rec.TrackedDownloadState,
				Size:                    rec.Size,
				SizeLeft:                rec.Sizeleft,
				EstimatedCompletionTime: rec.EstimatedCompletionTime,
			}, nil
		}
	}

	return types.DownloadingStatus{Found: false}, nil
}

/* Tools */

func toFilmStruct(film *radarr.Movie) Film {
	f := Film{
		TmdbId:        film.TmdbID,
		MovieId:       film.ID,
		IsInLibrary:   film.ID > 0,
		Title:         film.Title,
		OriginalTitle: film.OriginalTitle,
		Year:          film.Year,
		Runtime:       film.Runtime,
		Overview:      film.Overview,
		Genres:        film.Genres,
		Studio:        film.Studio,
		Size:          0,
	}

	if film.MovieFile != nil {
		f.Downloaded = true
		if film.MovieFile != nil {
			f.Quality = film.MovieFile.Quality.Quality.Name
			f.Size = float64(film.MovieFile.Size) / 1024 / 1024 / 1024
		}
	} else {
		f.Downloaded = false
	}

	// get the rating from the ratings list
	for src, rating := range film.Ratings {
		if src == "tmdb" {
			f.Rating += rating.Value
			f.NumberOfVotes += int(rating.Votes)
			f.RatingSrc = "TMDb"
			break
		} else if src == "imdb" {
			f.Rating += rating.Value
			f.NumberOfVotes += int(rating.Votes)
			f.RatingSrc = "IMDb"
		}
	}

	// get the cover image
	for _, image := range film.Images {
		if image.CoverType == "poster" {
			f.CoverImage = image.RemoteURL
			break
		}
	}

	return f
}
