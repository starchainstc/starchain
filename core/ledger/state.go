package ledger

type State struct {

}

type StateStore interface {
	SaveState(*State) error
}
