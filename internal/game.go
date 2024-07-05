package internal

type Game struct {
	currentPosition Position
	positions       []Position
}

func NewGame() *Game {
	return &Game{}
}

func (game *Game) Move(startIdx int, endIdx int) {
	//
}

func (game *Game) Revert() {

}
