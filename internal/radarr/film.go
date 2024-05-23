package radarr

import "strconv"

type Film struct {
	TmdbId  int64
	MovieId int64

	// IsInLibrary is true if the film is in the library.
	IsInLibrary bool

	// Title is the title of the film.
	Title string
	// OriginalTitle is the original title of the film.
	OriginalTitle string
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
	// RatingSrc is the source of the rating (e.g. "tmdb").
	RatingSrc string

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

func (f Film) PrintMovieTitle() string {
	str := "🎬 *" + f.Title + "* (_" + strconv.Itoa(f.Year) + "_)"

	return str
}

func (f Film) PrintMovieTitleAndInLibrary() string {
	str := f.PrintMovieTitle()
	if f.IsInLibrary {
		str += "\n\nAlready in library ✅"
	}

	return str
}

func (f Film) PrintMovieDetails() string {
	str := "📅 " + "*Release year*: " + strconv.Itoa(f.Year) + "\n"
	str += "🕒 " + "*Duration*: " + strconv.Itoa(f.Runtime) + " mins" + "\n"
	str += "⭐ " + "*Rating* (" + f.RatingSrc + "): " + strconv.FormatFloat(f.Rating, 'f', 1, 64) + "/10 (_" + strconv.Itoa(f.NumberOfVotes) + " votes_)\n"
	str += "🎞 " + "*Genres*: "
	for i, genre := range f.Genres {
		if i != 0 {
			str += ", "
		}
		str += genre
	}
	str += "\n"
	str += "🏢 " + "*Studio*: " + f.Studio + "\n"
	str += "📝 " + "*Overview*: "
	if len(f.Overview) > 175 {
		str += f.Overview[:175] + "..."
	} else {
		str += f.Overview
	}
	str += " [TMDb](https://www.themoviedb.org/movie/" + strconv.Itoa(int(f.TmdbId)) + ") 🔗\n\n"

	str += "📡 " + "*Status*: "
	if f.Downloaded {
		str += "Downloaded ✅\n"
		str += "📺 " + "*Quality*: " + f.Quality + "\n"
		str += "💾 " + "*Size*: " + strconv.FormatFloat(f.Size, 'f', 2, 64) + " GB\n"
	} else {
		str += "Missing ❌\n"
	}

	str += "\n"

	str += "[__View in Radarr__](nasalex.hole:30025/movie/" + strconv.Itoa(int(f.TmdbId)) + ")"

	return str
}

// ToFolderPath returns the path where to store the film.
// It does not return the current path of the film, but the path where to store it.
func (f Film) ToFolderPath() string {
	return "/movies/" + f.OriginalTitle + " (" + strconv.Itoa(f.Year) + ")"
}
