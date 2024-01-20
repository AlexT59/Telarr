package updates

type Action interface {
	isAction()
	String() string
}

type userAction string

func (a userAction) isAction() {}

func (a userAction) String() string {
	return string(a)
}

const (
	// userActionLookMovieToAdd is the action to look for a movie to add.
	userActionLookMovieToAdd userAction = "lookMovieToAdd"
	// movieDetails is the action to ask for the details of a movie.
	userActionMovieDetails userAction = "movieDetails"
	// removeMovie is the action to ask for the removal of a movie.
	userActionRemoveMovie userAction = "removeMovie"

	// lookSerieToAdd is the action to look for a serie to add.
	userActionLookSerieToAdd userAction = "lookSerieToAdd"
	// serieDetails is the action to ask for the details of a serie.
	userActionSerieDetails userAction = "serieDetails"
	// removeSerie is the action to ask for the removal of a serie.
	userActionRemoveSerie userAction = "removeSerie"
)

type callbackAction string

func (a callbackAction) isAction() {}

func (a callbackAction) String() string {
	return string(a)
}

const (
	// callbackNextMovie is the action to get the next page of the movies list.
	callbackNextMovie callbackAction = "nextMovie"
	// callbackPreviousMovie is the action to get the previous page of the movies list.
	callbackPreviousMovie callbackAction = "previousMovie"
	// callbackFirstMovie is the action to get the first page of the movies list.
	callbackFirstMovie callbackAction = "firstMovie"
	// callbackLastMovie is the action to get the last page of the movies list.
	callbackLastMovie callbackAction = "lastMovie"
	// callbackMovieDetails is the action to get the details of a movie.
	callbackMovieDetails callbackAction = "movieDetails"
	// callbackBackToMoviesList is the action to go back to the movies list, from the movie details.
	callbackBackToMoviesList callbackAction = "backToMoviesList"
	// callbackRemoveMovie is the action to remove a movie.
	callbackRemoveMovie callbackAction = "removeMovie"
	// callbackConfirmRemoveMovie is the action to confirm the removal of a movie.
	callbackConfirmRemoveMovie callbackAction = "confirmRemoveMovie"
	// callbackCancelRemoveMovie is the action to cancel the removal of a movie.
	callbackCancelRemoveMovie callbackAction = "cancelRemoveMovie"
	// callbackNextAddMovie is the action to get the movie, from the list of movies to add.
	callbackNextAddMovie callbackAction = "nextAddMovie"
	// callbackPreviousAddMovie is the action to get the movie, from the list of movies to add.
	callbackPreviousAddMovie callbackAction = "previousAddMovie"
	// callbackEditRequestAddMovie is the action to edit the request of add a movie.
	callbackEditRequestAddMovie callbackAction = "editRequestMovie"

	// callbackNextSerie is the action to get the next page of the series list.
	callbackNextSerie callbackAction = "nextSerie"
	// callbackPreviousSerie is the action to get the previous page of the series list.
	callbackPreviousSerie callbackAction = "previousSerie"
	// callbackFirstSerie is the action to get the first page of the series list.
	callbackFirstSerie callbackAction = "firstSerie"
	// callbackLastSerie is the action to get the last page of the series list.
	callbackLastSerie callbackAction = "lastSerie"
	// callbackSerieDetails is the action to get the details of a serie.
	callbackSerieDetails callbackAction = "serieDetails"
	// callbackBackToSerieList is the action to go back to the series list, from the serie details.
	callbackBackToSeriesList callbackAction = "backToSeriesList"
	// callbackRemoveSerie is the action to remove a serie.
	callbackRemoveSerie callbackAction = "removeSerie"
	// callbackConfirmRemoveSerie is the action to confirm the removal of a serie.
	callbackConfirmRemoveSerie callbackAction = "confirmRemoveSerie"
	// callbackCancelRemoveSerie is the action to cancel the removal of a serie.
	callbackCancelRemoveSerie callbackAction = "cancelRemoveSerie"
	// callbackNextAddSerie is the action to get the serie, from the list of series to add.
	callbackNextAddSerie callbackAction = "nextAddSerie"
	// callbackPreviousAddSerie is the action to get the serie, from the list of series to add.
	callbackPreviousAddSerie callbackAction = "previousAddSerie"
	// callbackEditRequestAddSerie is the action to edit the request of add a serie.
	callbackEditRequestAddSerie callbackAction = "editRequestSerie"

	// callbackCancel is the action to cancel the current action.
	callbackCancel callbackAction = "cancel"
)
