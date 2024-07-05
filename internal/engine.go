package internal

type MoveGenerator interface {
}

type Engine struct {
	game *Game
}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) StartGame() {}

func (e *Engine) Move() {}

func (e *Engine) GetPossibleMoves(pieceIdx int) {}
