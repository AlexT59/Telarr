package radarr

import (
	"strconv"
	"telarr/configuration"

	"github.com/rs/zerolog/log"
	"golift.io/starr"
	"golift.io/starr/radarr"
)

type Film struct {
	TmdbId  int64
	MovieId int64

	// Title is the title of the film.
	Title string
	// Year is the release year of the film.
	Year int

	// CoverImage is the url to the cover of the film.
	CoverImage string

	// Runtime is the runtime of the film in minutes.
	Runtime int

	// Rating is the rating of the film (/10).
	Rating float64
	// NumberOfVotes is the number of votes for the film.
	NumberOfVotes int

	// Overview is the overview of the film.
	Overview string
	// Genres is the list of genres of the film.
	Genres []string
	// Studio is the studio of the film.
	Studio string

	// Quality is the quality of the film.
	Quality string
	// Downloaded is true if the film is downloaded.
	Downloaded bool
	// Size is the size of the film on disk GB.
	Size float64
}

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

/* Film */

func (f Film) PrintMovieTitle() string {
	str := "🎬 *" + f.Title + "* (_" + strconv.Itoa(f.Year) + "_)"

	return str
}

func (f Film) PrintMovieDetails() string {
	str := "📅 " + "*Release year*: " + strconv.Itoa(f.Year) + "\n"
	str += "🕒 " + "*Duration*: " + strconv.Itoa(f.Runtime) + " mins" + "\n"
	str += "⭐ " + "*Rating*: " + strconv.FormatFloat(f.Rating, 'f', 1, 64) + "/10 (_" + strconv.Itoa(f.NumberOfVotes) + " votes_)\n"
	str += "🎞 " + "*Genres*: "
	for i, genre := range f.Genres {
		if i != 0 {
			str += ", "
		}
		str += genre
	}
	str += "\n"
	str += "🏢 " + "*Studio*: " + f.Studio + "\n"
	str += "📝 " + "*Overview*: " + f.Overview + "\n"
	str += "\n"

	str += "📡 " + "*Status*: "
	if f.Downloaded {
		str += "Downloaded ✅\n"
		str += "📺 " + "*Quality*: " + f.Quality + "\n"
		str += "💾 " + "*Size*: " + strconv.FormatFloat(f.Size, 'f', 2, 64) + " GB\n"
	} else {
		str += "Missing ❌\n"
	}

	str += "\n🔗 [The Movie DB](https://www.themoviedb.org/movie/" + strconv.Itoa(int(f.TmdbId)) + ")"

	return str
}

/* Tools */

func toFilmStruct(film *radarr.Movie) Film {
	f := Film{
		TmdbId:   film.TmdbID,
		MovieId:  film.ID,
		Title:    film.Title,
		Year:     film.Year,
		Runtime:  film.Runtime,
		Overview: film.Overview,
		Genres:   film.Genres,
		Studio:   film.Studio,
		Size:     0,
	}

	if film.HasFile {
		f.Downloaded = true
		if film.MovieFile != nil {
			f.Quality = film.MovieFile.Quality.Quality.Name
			f.Size = float64(film.MovieFile.Size) / 1024 / 1024 / 1024
		}
	} else {
		f.Downloaded = false
	}

	// get the rating from the ratings list
	r := 0.0
	nbSources := 0
	for _, rating := range film.Ratings {
		if rating.Type == "user" {
			if rating.Value > 10 {
				// value in %, convert it to /10
				rating.Value /= 10
			}
			r += rating.Value
			nbSources++
			f.NumberOfVotes += int(rating.Votes)
			break
		}
	}
	f.Rating = r / float64(nbSources)

	// get the cover image
	for _, image := range film.Images {
		if image.CoverType == "poster" {
			f.CoverImage = image.RemoteURL
			break
		}
	}

	return f
}
