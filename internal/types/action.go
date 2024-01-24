package types

type Action interface {
	isAction()
	String() string
}

type UserAction string

func (a UserAction) isAction() {}

func (a UserAction) String() string {
	return string(a)
}

const (
	// UserActionLookMovieToAdd is the action to look for a movie to add.
	UserActionLookMovieToAdd UserAction = "lookMovieToAdd"
	// UserActionMovieDetails is the action to ask for the details of a movie.
	UserActionMovieDetails UserAction = "movieDetails"
	// UserActionRemoveMovie is the action to ask for the removal of a movie.
	UserActionRemoveMovie UserAction = "removeMovie"
	// UserActionAddMovie is the action to ask for the addition of a movie.
	UserActionAddMovie UserAction = "addMovie"

	// UserActionlookSerieToAdd is the action to look for a serie to add.
	UserActionLookSerieToAdd UserAction = "lookSerieToAdd"
	// UserActionSerieDetails is the action to ask for the details of a serie.
	UserActionSerieDetails UserAction = "serieDetails"
	// UserActionRemoveSerie is the action to ask for the removal of a serie.
	UserActionRemoveSerie UserAction = "removeSerie"
	// UserActionAddSerie is the action to ask for the addition of a serie.
	UserActionAddSerie UserAction = "addSerie"
)

type CallbackAction string

func (a CallbackAction) isAction() {}

func (a CallbackAction) String() string {
	return string(a)
}

const (
	// CallbackNextMovie is the action to get the next page of the movies list.
	CallbackNextMovie CallbackAction = "nextMovie"
	// CallbackPreviousMovie is the action to get the previous page of the movies list.
	CallbackPreviousMovie CallbackAction = "previousMovie"
	// CallbackFirstMovie is the action to get the first page of the movies list.
	CallbackFirstMovie CallbackAction = "firstMovie"
	// CallbackLastMovie is the action to get the last page of the movies list.
	CallbackLastMovie CallbackAction = "lastMovie"
	// CallbackMovieDetails is the action to get the details of a movie.
	CallbackMovieDetails CallbackAction = "movieDetails"
	// CallbackBackToMoviesList is the action to go back to the movies list, from the movie details.
	CallbackBackToMoviesList CallbackAction = "backToMoviesList"
	// CallbackRemoveMovie is the action to remove a movie.
	CallbackRemoveMovie CallbackAction = "removeMovie"
	// CallbackConfirmRemoveMovie is the action to confirm the removal of a movie.
	CallbackConfirmRemoveMovie CallbackAction = "confirmRemoveMovie"
	// CallbackCancelRemoveMovie is the action to cancel the removal of a movie.
	CallbackCancelRemoveMovie CallbackAction = "cancelRemoveMovie"
	// CallbackAddMovie is the action to add a movie.
	CallbackAddMovie CallbackAction = "addMovie"
	// CallbackNextAddMovie is the action to get the movie, from the list of movies to add.
	CallbackNextAddMovie CallbackAction = "nextAddMovie"
	// CallbackPreviousAddMovie is the action to get the movie, from the list of movies to add.
	CallbackPreviousAddMovie CallbackAction = "previousAddMovie"
	// CallbackEditRequestAddMovie is the action to edit the request of add a movie.
	CallbackEditRequestAddMovie CallbackAction = "editRequestMovie"

	// CallbackNextSerie is the action to get the next page of the series list.
	CallbackNextSerie CallbackAction = "nextSerie"
	// CallbackPreviousSerie is the action to get the previous page of the series list.
	CallbackPreviousSerie CallbackAction = "previousSerie"
	// CallbackFirstSerie is the action to get the first page of the series list.
	CallbackFirstSerie CallbackAction = "firstSerie"
	// CallbackLastSerie is the action to get the last page of the series list.
	CallbackLastSerie CallbackAction = "lastSerie"
	// CallbackSerieDetails is the action to get the details of a serie.
	CallbackSerieDetails CallbackAction = "serieDetails"
	// CallbackBackToSerieList is the action to go back to the series list, from the serie details.
	CallbackBackToSeriesList CallbackAction = "backToSeriesList"
	// CallbackRemoveSerie is the action to remove a serie.
	CallbackRemoveSerie CallbackAction = "removeSerie"
	// CallbackConfirmRemoveSerie is the action to confirm the removal of a serie.
	CallbackConfirmRemoveSerie CallbackAction = "confirmRemoveSerie"
	// CallbackCancelRemoveSerie is the action to cancel the removal of a serie.
	CallbackCancelRemoveSerie CallbackAction = "cancelRemoveSerie"
	// CallbackAddSerie is the action to add a serie.
	CallbackAddSerie CallbackAction = "addSerie"
	// CallbackNextAddSerie is the action to get the serie, from the list of series to add.
	CallbackNextAddSerie CallbackAction = "nextAddSerie"
	// CallbackPreviousAddSerie is the action to get the serie, from the list of series to add.
	CallbackPreviousAddSerie CallbackAction = "previousAddSerie"
	// CallbackEditRequestAddSerie is the action to edit the request of add a serie.
	CallbackEditRequestAddSerie CallbackAction = "editRequestSerie"

	// CallbackCancel is the action to cancel the current action.
	CallbackCancel CallbackAction = "cancel"
)
