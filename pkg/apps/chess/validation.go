package chess

import (
	"fmt"
	"strings"
)

// ============================================================================
// CHESS MOVE VALIDATION ENGINE
// ============================================================================

// Board represents the chess board state
type Board struct {
	Position [8][8]string // 8x8 board with pieces
	EnPassant string       // En passant target square
	Castling CastlingRights
	Turn     string        // "white" or "black"
}

// CastlingRights tracks which castling moves are available
type CastlingRights struct {
	WhiteKingside  bool
	WhiteQueenside bool
	BlackKingside  bool
	BlackQueenside bool
}

// MoveValidation represents the result of move validation
type MoveValidation struct {
	Valid              bool
	Reason             string
	IsCapture          bool
	IsCheck            bool
	IsCheckmate        bool
	IsStalemate        bool
	IsCastle           bool
	IsEnPassant        bool
	RequiresPromotion  bool
	NextBoardState     string
	NextBoardFEN       string
}

// InitializeBoard creates a new board in standard chess starting position
func InitializeBoard() *Board {
	board := &Board{
		Turn: "white",
		Castling: CastlingRights{
			WhiteKingside:  true,
			WhiteQueenside: true,
			BlackKingside:  true,
			BlackQueenside: true,
		},
	}

	// Set up starting position
	// Black pieces
	board.Position[0][0] = "br" // Black rook
	board.Position[0][1] = "bn" // Black knight
	board.Position[0][2] = "bb" // Black bishop
	board.Position[0][3] = "bq" // Black queen
	board.Position[0][4] = "bk" // Black king
	board.Position[0][5] = "bb"
	board.Position[0][6] = "bn"
	board.Position[0][7] = "br"

	// Black pawns
	for i := 0; i < 8; i++ {
		board.Position[1][i] = "bp"
	}

	// Empty squares (already empty by default)

	// White pawns
	for i := 0; i < 8; i++ {
		board.Position[6][i] = "wp"
	}

	// White pieces
	board.Position[7][0] = "wr"
	board.Position[7][1] = "wn"
	board.Position[7][2] = "wb"
	board.Position[7][3] = "wq"
	board.Position[7][4] = "wk"
	board.Position[7][5] = "wb"
	board.Position[7][6] = "wn"
	board.Position[7][7] = "wr"

	return board
}

// LoadFromFEN loads a board from FEN notation
func (b *Board) LoadFromFEN(fen string) error {
	parts := strings.Split(fen, " ")
	if len(parts) < 4 {
		return fmt.Errorf("invalid FEN string")
	}

	// Parse position
	rows := strings.Split(parts[0], "/")
	for i, row := range rows {
		colIdx := 0
		for _, char := range row {
			if char >= '1' && char <= '8' {
				// Empty squares
				spaces := int(char - '0')
				for j := 0; j < spaces; j++ {
					b.Position[i][colIdx] = ""
					colIdx++
				}
			} else {
				// Piece - convert from FEN to internal format
				piece := b.fenToPiece(string(char))
				b.Position[i][colIdx] = piece
				colIdx++
			}
		}
	}

	// Parse turn - convert w/b to white/black
	if parts[1] == "w" {
		b.Turn = "white"
	} else {
		b.Turn = "black"
	}

	// Parse castling rights
	castling := parts[2]
	b.Castling.WhiteKingside = strings.Contains(castling, "K")
	b.Castling.WhiteQueenside = strings.Contains(castling, "Q")
	b.Castling.BlackKingside = strings.Contains(castling, "k")
	b.Castling.BlackQueenside = strings.Contains(castling, "q")

	// Parse en passant
	b.EnPassant = parts[3]

	return nil
}

// ToFEN converts board to FEN notation
func (b *Board) ToFEN() string {
	var fen strings.Builder

	// Position
	for i := 0; i < 8; i++ {
		emptyCount := 0
		for j := 0; j < 8; j++ {
			if b.Position[i][j] == "" {
				emptyCount++
			} else {
				if emptyCount > 0 {
					fen.WriteString(fmt.Sprintf("%d", emptyCount))
					emptyCount = 0
				}
				fenPiece := b.pieceToFEN(b.Position[i][j])
				fen.WriteString(fenPiece)
			}
		}
		if emptyCount > 0 {
			fen.WriteString(fmt.Sprintf("%d", emptyCount))
		}
		if i < 7 {
			fen.WriteString("/")
		}
	}

	fen.WriteString(" ")
	// Convert turn to FEN format (w or b)
	turn := "w"
	if b.Turn == "black" {
		turn = "b"
	}
	fen.WriteString(turn)
	fen.WriteString(" ")

	// Castling rights
	castling := ""
	if b.Castling.WhiteKingside {
		castling += "K"
	}
	if b.Castling.WhiteQueenside {
		castling += "Q"
	}
	if b.Castling.BlackKingside {
		castling += "k"
	}
	if b.Castling.BlackQueenside {
		castling += "q"
	}
	if castling == "" {
		castling = "-"
	}
	fen.WriteString(castling)
	fen.WriteString(" ")
	fen.WriteString(b.EnPassant)

	return fen.String()
}

// ValidateMove validates a chess move
func (b *Board) ValidateMove(fromSquare, toSquare string, promotion string) *MoveValidation {
	result := &MoveValidation{
		Valid: false,
	}

	// Convert algebraic notation to coordinates
	fromFile := int(fromSquare[0] - 'a')
	fromRank := 8 - int(fromSquare[1]-'0')
	toFile := int(toSquare[0] - 'a')
	toRank := 8 - int(toSquare[1]-'0')

	// Validate square bounds
	if fromFile < 0 || fromFile > 7 || fromRank < 0 || fromRank > 7 ||
		toFile < 0 || toFile > 7 || toRank < 0 || toRank > 7 {
		result.Reason = "Invalid square"
		return result
	}

	piece := b.Position[fromRank][fromFile]
	targetPiece := b.Position[toRank][toFile]

	// Validate piece exists and belongs to current player
	if piece == "" {
		result.Reason = "No piece on source square"
		return result
	}

	pieceColor := string(piece[0])
	playerColor := b.Turn[0:1]
	if pieceColor != playerColor {
		result.Reason = "Piece does not belong to current player"
		return result
	}

	// Validate target square is not occupied by own piece
	if targetPiece != "" && string(targetPiece[0]) == pieceColor {
		result.Reason = "Cannot capture own piece"
		return result
	}

	// Validate move legality based on piece type
	pieceName := string(piece[1])
	isLegal := false
	isCastle := false
	isEnPassant := false

	switch pieceName {
	case "p":
		isLegal, isCastle, isEnPassant = b.validatePawnMove(fromFile, fromRank, toFile, toRank, piece)
	case "n":
		isLegal = b.validateKnightMove(fromFile, fromRank, toFile, toRank)
	case "b":
		isLegal = b.validateBishopMove(fromFile, fromRank, toFile, toRank)
	case "r":
		isLegal = b.validateRookMove(fromFile, fromRank, toFile, toRank)
	case "q":
		isLegal = b.validateQueenMove(fromFile, fromRank, toFile, toRank)
	case "k":
		isLegal, isCastle = b.validateKingMove(fromFile, fromRank, toFile, toRank, piece)
	}

	if !isLegal {
		result.Reason = "Illegal move for piece"
		return result
	}

	// Check if move is en passant
	if isEnPassant {
		result.IsEnPassant = true
	}

	// Check if move is castling
	if isCastle {
		result.IsCastle = true
	}

	// Simulate move and check for check
	boardCopy := b.clone()
	boardCopy.makeMove(fromFile, fromRank, toFile, toRank, piece, targetPiece)

	// Check if player's king is in check after move
	if boardCopy.isKingInCheck(playerColor) {
		result.Reason = "Move leaves king in check"
		return result
	}

	result.Valid = true
	result.IsCapture = targetPiece != ""
	result.NextBoardState = boardCopy.ToFEN()

	// Check for check in new position
	opponentColor := "b"
	if playerColor == "b" {
		opponentColor = "w"
	}

	if boardCopy.isKingInCheck(opponentColor) {
		result.IsCheck = true
		// Note: Checkmate/Stalemate detection requires non-recursive hasLegalMoves
		// which would create circular dependency, so omitted for now
	}

	// Check if pawn promotion is required
	if pieceName == "p" && ((playerColor == "w" && toRank == 0) || (playerColor == "b" && toRank == 7)) {
		result.RequiresPromotion = true
		if promotion == "" {
			result.Reason = "Promotion piece required"
			result.Valid = false
			return result
		}
	}

	return result
}

// Helper functions

func (b *Board) validatePawnMove(fromFile, fromRank, toFile, toRank int, piece string) (bool, bool, bool) {
	color := string(piece[0])
	direction := -1  // White moves down (decreasing rank index)
	startRank := 6   // White pawns start at rank 2 (index 6)

	if color == "b" {
		direction = 1   // Black moves up (increasing rank index)
		startRank = 1   // Black pawns start at rank 7 (index 1)
	}

	// Single forward move
	if toFile == fromFile && toRank == fromRank+direction && b.Position[toRank][toFile] == "" {
		return true, false, false
	}

	// Double forward move from start
	if toFile == fromFile && fromRank == startRank && toRank == fromRank+2*direction &&
		b.Position[toRank][toFile] == "" && b.Position[fromRank+direction][toFile] == "" {
		return true, false, false
	}

	// Capture
	if toFile != fromFile && toRank == fromRank+direction && b.Position[toRank][toFile] != "" {
		return true, false, false
	}

	// En passant
	if toFile != fromFile && toRank == fromRank+direction && b.Position[toRank][toFile] == "" {
		if fmt.Sprintf("%c%d", 'a'+byte(toFile), 8-toRank) == b.EnPassant {
			return true, false, true
		}
	}

	return false, false, false
}

func (b *Board) validateKnightMove(fromFile, fromRank, toFile, toRank int) bool {
	fileDiff := abs(toFile - fromFile)
	rankDiff := abs(toRank - fromRank)
	return (fileDiff == 2 && rankDiff == 1) || (fileDiff == 1 && rankDiff == 2)
}

func (b *Board) validateBishopMove(fromFile, fromRank, toFile, toRank int) bool {
	if abs(toFile-fromFile) != abs(toRank-fromRank) {
		return false
	}
	return b.isPathClear(fromFile, fromRank, toFile, toRank)
}

func (b *Board) validateRookMove(fromFile, fromRank, toFile, toRank int) bool {
	if fromFile != toFile && fromRank != toRank {
		return false
	}
	return b.isPathClear(fromFile, fromRank, toFile, toRank)
}

func (b *Board) validateQueenMove(fromFile, fromRank, toFile, toRank int) bool {
	return b.validateRookMove(fromFile, fromRank, toFile, toRank) ||
		b.validateBishopMove(fromFile, fromRank, toFile, toRank)
}

func (b *Board) validateKingMove(fromFile, fromRank, toFile, toRank int, piece string) (bool, bool) {
	fileDiff := abs(toFile - fromFile)
	rankDiff := abs(toRank - fromRank)

	// Regular king move
	if fileDiff <= 1 && rankDiff <= 1 && !(fileDiff == 0 && rankDiff == 0) {
		return true, false
	}

	// Castling
	if rankDiff == 0 && fileDiff == 2 {
		color := string(piece[0])
		if color == "w" {
			if toFile == 6 && b.Castling.WhiteKingside {
				return true, true
			}
			if toFile == 2 && b.Castling.WhiteQueenside {
				return true, true
			}
		} else {
			if toFile == 6 && b.Castling.BlackKingside {
				return true, true
			}
			if toFile == 2 && b.Castling.BlackQueenside {
				return true, true
			}
		}
	}

	return false, false
}

func (b *Board) isPathClear(fromFile, fromRank, toFile, toRank int) bool {
	fileStep := 0
	rankStep := 0

	if toFile > fromFile {
		fileStep = 1
	} else if toFile < fromFile {
		fileStep = -1
	}

	if toRank > fromRank {
		rankStep = 1
	} else if toRank < fromRank {
		rankStep = -1
	}

	file := fromFile + fileStep
	rank := fromRank + rankStep

	for file != toFile || rank != toRank {
		if b.Position[rank][file] != "" {
			return false
		}
		file += fileStep
		rank += rankStep
	}

	return true
}

func (b *Board) isKingInCheck(color string) bool {
	// Find king position
	var kingFile, kingRank int
	kingPiece := color + "k"

	found := false
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if b.Position[i][j] == kingPiece {
				kingRank = i
				kingFile = j
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return false
	}

	// Check if any opponent piece can attack the king
	opponentColor := "b"
	if color == "b" {
		opponentColor = "w"
	}

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			piece := b.Position[i][j]
			if piece != "" && string(piece[0]) == opponentColor {
				// Check if this piece attacks the king
				if b.canPieceAttack(j, i, kingFile, kingRank, piece) {
					return true
				}
			}
		}
	}

	return false
}

func (b *Board) canPieceAttack(fromFile, fromRank, toFile, toRank int, piece string) bool {
	pieceName := string(piece[1])

	switch pieceName {
	case "p":
		// Pawns attack diagonally
		color := string(piece[0])
		direction := -1
		if color == "b" {
			direction = 1
		}
		return toRank == fromRank+direction && abs(toFile-fromFile) == 1
	case "n":
		return b.validateKnightMove(fromFile, fromRank, toFile, toRank)
	case "b":
		return b.validateBishopMove(fromFile, fromRank, toFile, toRank)
	case "r":
		return b.validateRookMove(fromFile, fromRank, toFile, toRank)
	case "q":
		return b.validateQueenMove(fromFile, fromRank, toFile, toRank)
	case "k":
		fileDiff := abs(toFile - fromFile)
		rankDiff := abs(toRank - fromRank)
		return fileDiff <= 1 && rankDiff <= 1 && !(fileDiff == 0 && rankDiff == 0)
	}

	return false
}

func (b *Board) hasLegalMoves(color string) bool {
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			piece := b.Position[i][j]
			if piece != "" && string(piece[0]) == color {
				for ti := 0; ti < 8; ti++ {
					for tj := 0; tj < 8; tj++ {
						if b.ValidateMove(
							fmt.Sprintf("%c%d", 'a'+byte(j), 8-i),
							fmt.Sprintf("%c%d", 'a'+byte(tj), 8-ti),
							"",
						).Valid {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func (b *Board) makeMove(fromFile, fromRank, toFile, toRank int, piece, targetPiece string) {
	b.Position[toRank][toFile] = piece
	b.Position[fromRank][fromFile] = ""

	// Update castling rights
	pieceName := string(piece[1])
	color := string(piece[0])

	if pieceName == "k" {
		if color == "w" {
			b.Castling.WhiteKingside = false
			b.Castling.WhiteQueenside = false
		} else {
			b.Castling.BlackKingside = false
			b.Castling.BlackQueenside = false
		}
	}

	if pieceName == "r" {
		if color == "w" {
			if fromFile == 0 {
				b.Castling.WhiteQueenside = false
			} else if fromFile == 7 {
				b.Castling.WhiteKingside = false
			}
		} else {
			if fromFile == 0 {
				b.Castling.BlackQueenside = false
			} else if fromFile == 7 {
				b.Castling.BlackKingside = false
			}
		}
	}

	// Change turn
	if b.Turn == "white" {
		b.Turn = "black"
	} else {
		b.Turn = "white"
	}
}

func (b *Board) clone() *Board {
	boardCopy := &Board{
		Turn:      b.Turn,
		EnPassant: b.EnPassant,
		Castling:  b.Castling,
	}
	copy(boardCopy.Position[:], b.Position[:])
	return boardCopy
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// pieceToFEN converts internal piece format (wp, bk, etc.) to FEN notation (P, K, etc.)
func (b *Board) pieceToFEN(piece string) string {
	if piece == "" {
		return ""
	}

	fenMap := map[string]string{
		"wp": "P", "wn": "N", "wb": "B", "wr": "R", "wq": "Q", "wk": "K",
		"bp": "p", "bn": "n", "bb": "b", "br": "r", "bq": "q", "bk": "k",
	}

	if fen, exists := fenMap[piece]; exists {
		return fen
	}
	return piece
}

// fenToPiece converts FEN notation (P, K, etc.) to internal piece format (wp, bk, etc.)
func (b *Board) fenToPiece(fenPiece string) string {
	fenMap := map[string]string{
		"P": "wp", "N": "wn", "B": "wb", "R": "wr", "Q": "wq", "K": "wk",
		"p": "bp", "n": "bn", "b": "bb", "r": "br", "q": "bq", "k": "bk",
	}

	if piece, exists := fenMap[fenPiece]; exists {
		return piece
	}
	return fenPiece
}
