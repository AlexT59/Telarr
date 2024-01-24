package sonarr

import "strconv"

type Serie struct {
	TvdbId  int64
	SerieId int64

	// IsInLibrary is true if the serie is in the library.
	IsInLibrary bool

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

	// Downloaded is true if the serie is downloaded.
	Downloaded bool
	// Size is the size of the serie on disk GB.
	Size float64
}

func (s Serie) PrintSerieTitle() string {
	str := "📺 *" + s.Title + "* (_" + strconv.Itoa(s.Year) + "_)"

	return str
}

func (s Serie) PrintSerieTitleAndInLibrary() string {
	str := s.PrintSerieTitle()
	if s.IsInLibrary {
		str += "\n\nAlready in library ✅"
	}

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

func (s Serie) ToFolderPath() string {
	return "/tv/" + s.Title + " (" + strconv.Itoa(s.Year) + ")"
}
