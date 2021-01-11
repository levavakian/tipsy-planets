package main

import (
	"math/rand"
	"sync"
	"time"
	"fmt"
	"log"
	"github.com/gorilla/websocket"
	"errors"
)

const (
	MOVE = "MOVE"
	BATTLE = "BATTLE"
	VICTORY = "VICTORY"
)

const (
	DICE_SIZE = 6
)

type Settings struct {
	RequireExactVictory bool `json:"require_exact_victory"`
}

type Player struct {
	Name string `json:"name"`
	Location string `json:"location"`
	Conns map[*websocket.Conn]bool `json:"-"`
}

type LocationEffect struct {
	IsWormhole bool `json:"is_wormhole"`
	WormhomeTarget string `json:"wormhole_target"`
}

type Location struct {
	Name string `json:"name"`
	X int `json:"x"`
	Y int `json:"y"`
	Effects []*LocationEffect `json:"effects"`
}

type GameBoard struct {
	Locations []*Location `json:"locations"`
}

type Input struct {
	Name string `json:"name"`
	Value int `json:"value"`
	Code string `json:"code"`
}

type InputRequest struct {
	Names []string `json:"names"`
	Type string `json:"type"`
	Received []*Input `json:"received"`
}

func (i *InputRequest) GetReceivedForName(name string) *Input {
	for _, rec := range i.Received {
		if rec.Name == name {
			return rec
		}
	}
	return nil
}

type Room struct {
	sync.RWMutex
	Code string `json:"code"`
	Players []*Player `json:"players"`
	CurrentPlayer string `json:"current_player"`
	Board GameBoard `json:"board"`
	LastUpdate time.Time `json:"last_update"`
	InputReqs []*InputRequest `json:"input_reqs"`
	History []string `json:"history"`
	Settings Settings `json:"settings"`
}

func newRoom(code string) *Room {
	return &Room{
		Code: code,
		Players: []*Player{},
		Board: defaultGameBoard(),
		LastUpdate: time.Now(),
		InputReqs: []*InputRequest{},
		History: []string{},
		Settings: Settings{
			RequireExactVictory: false,
		},
	}
}

func (r *Room) PopInputReq() {
	r.InputReqs[0] = nil
	r.InputReqs = r.InputReqs[1:]
}

func (r *Room) ClearPendingBattlesForPlayer(name string) {
	nInputReqs := []*InputRequest{}
	for _, req := range r.InputReqs {
		hasPlayer := func()bool{
			if len(req.Names) == len(req.Received) {
				return false
			}
			for _, candidate := range req.Names {
				if candidate == name {
					return true
				}
			}
			return false
		}()
		if !hasPlayer {
			nInputReqs = append(nInputReqs, req)
		}
	}
	r.InputReqs = nInputReqs
}

func defaultGameBoard() GameBoard {
	locs := []*Location{
		{"1", 50, 50, []*LocationEffect{}},
		{"2", 200, 50, []*LocationEffect{{IsWormhole: true, WormhomeTarget: "C"}}},
		{"3", 450, 50, []*LocationEffect{}},
		{"4", 450, 200, []*LocationEffect{}},
		{"5", 200, 200, []*LocationEffect{}},
		{"6", 50, 200, []*LocationEffect{}},
		{"7", 50, 450, []*LocationEffect{}},
		{"8", 200, 450, []*LocationEffect{}},
		{"9", 450, 450, []*LocationEffect{}},
	}
	return GameBoard{
		Locations: locs,
	}
}

func getIdx(s []string, e string) (int, bool) {
    for idx, a := range s {
        if a == e {
            return idx, true
        }
    }
    return -1, false
}

func (r *Room) GetPlayer(name string) (*Player, int) {
	for idx, player := range r.Players {
		if player.Name == name {
			return player, idx
		}
	}
	return nil, 0
}

func (b *GameBoard) GetLocation(name string) (*Location, int) {
	for idx, loc := range b.Locations {
		if loc.Name == name {
			return loc, idx
		}
	}
	return nil, 0
}

func (r *Room) DoMove(input *InputRequest) error {
	// Return if we don't have all the inputs we're waiting for
	if len(input.Received) != len(input.Names) {
		return nil
	}
		
	dice := rand.Intn(6) + 1
	err := r.MovePlayer(input.Received[0].Name, dice, false)
	if err != nil {
		return err
	}

	r.PopInputReq()
	return nil
}

func (r *Room) DoVictory(input *InputRequest) error {
	// Return if we don't have all the inputs we're waiting for
	if len(input.Received) != len(input.Names) {
		return nil
	}

	r.PopInputReq()
	return nil
}

func (r *Room)MovePlayer(name string, amount int, forced bool) error {
	player, _ := r.GetPlayer(name)
	if player == nil {
		return errors.New("player not found")
	}

	_, lidx := r.Board.GetLocation(player.Location)

	newLocIdx := lidx + amount
	lastIdx := len(r.Board.Locations) - 1
	if newLocIdx > lastIdx {
		if r.Settings.RequireExactVictory {
			newLocIdx = lastIdx - (newLocIdx - lastIdx)
		} else {
			newLocIdx = lastIdx
		}
	}
	if newLocIdx < 0 {
		newLocIdx = 0
	}
	newLocIdx = newLocIdx % len(r.Board.Locations)
	newLoc := r.Board.Locations[newLocIdx].Name
	if !forced {
		r.History = append(r.History,
			fmt.Sprintf("%s rolled a %d and moved from %s to %s", player.Name, amount, player.Location, newLoc))
	} else {
		r.History = append(r.History,
			fmt.Sprintf("%s moved from %s to %s", player.Name, player.Location, newLoc))
	}
	player.Location = newLoc

	// Check if any other players are at the target location and set up battles if they are
	for _,  other := range r.Players {
		if other.Location == player.Location && other.Name != player.Name {
			r.InputReqs = append(r.InputReqs, &InputRequest{
				Type: BATTLE,
				Names: []string{player.Name, other.Name},
				Received: []*Input{},
			})
		}
	}
	return nil
}

func (r *Room) DoBattle(input *InputRequest) error {
	lastInput := input.Received[len(input.Received)-1]
	lastInput.Value = rand.Intn(6) + 1
	r.History = append(r.History, fmt.Sprintf("%s rolled a %d!", lastInput.Name, lastInput.Value))

	// Return if we don't have all the inputs we're waiting for
	if len(input.Received) != len(input.Names) {
		return nil
	}

	// Panic if the players are not in the same place, invalid game state
	playerOne, _ := r.GetPlayer(input.Names[0])
	playerTwo, _ := r.GetPlayer(input.Names[1])

	if (playerOne == nil || playerTwo == nil || playerOne.Location != playerTwo.Location) {
		log.Fatalln("Invalid game state during battle with", playerOne, playerTwo)
	}

	rollOne := input.GetReceivedForName(playerOne.Name).Value
	rollTwo := input.GetReceivedForName(playerTwo.Name).Value


	_, loser, diff := func()(*Player, *Player, int){
		if rollOne > rollTwo {
			return playerOne, playerTwo, rollTwo - rollOne
		} else {
			return playerTwo, playerOne, rollOne - rollTwo
		}
	}()

	if diff == 0 {
		r.History = append(r.History, "Battle was a tie!")
		r.PopInputReq()
		r.InputReqs = append([]*InputRequest{&InputRequest{
			Names: []string{playerOne.Name, playerTwo.Name},
			Type: BATTLE,
			Received: []*Input{},
		}}, r.InputReqs...)
		return nil
	}

	r.ClearPendingBattlesForPlayer(loser.Name)
	r.MovePlayer(loser.Name, diff, true)
	
	r.PopInputReq()
	return nil
}

func (r *Room) AdvanceRoomState(input *Input) (bool, error) {
	// Bail out if no one is playing
	if len(r.Players) == 0 {
		return false, errors.New("empty lobby")
	}

	// Bail out if we're starting a new game
	if len(r.InputReqs) == 0 {
		r.InputReqs = append(r.InputReqs, &InputRequest{
			Names: []string{r.Players[0].Name},
			Received: []*Input{},
			Type: MOVE,
		})
		r.History = append(r.History, input.Name + " started a new game")
		r.CurrentPlayer = r.Players[0].Name
		r.LastUpdate = time.Now()

		for _, player := range r.Players {
			player.Location = r.Board.Locations[0].Name
		}
		return true, nil
	}

	inputReq := r.InputReqs[0]

	// Bail out if we're not waiting for input from this player
	_, ok := getIdx(inputReq.Names, input.Name)
	if !ok {
		return false, errors.New("not your turn")
	}

	// Bail out if we already have input from this player
	for _, rec := range inputReq.Received {
		if rec.Name == input.Name {
			return false, errors.New("already have your input")
		}
	}

	// Add the input otherwise
	inputReq.Received = append(inputReq.Received, input)
	defer func(){
		r.LastUpdate = time.Now()
	}()

	// Now do the appropriate action
	var err error
	switch inputReq.Type {
	case MOVE:
		err = r.DoMove(inputReq)
	case BATTLE:
		err = r.DoBattle(inputReq)
	case VICTORY:
		err = r.DoVictory(inputReq)
		return true, err
	}

	// Do win conditions here
	for _, player := range r.Players {
		if player.Location == r.Board.Locations[len(r.Board.Locations) - 1].Name {
			r.History = append(r.History, fmt.Sprintf("%s won the round!", player.Name))
			r.InputReqs = []*InputRequest{&InputRequest{
				Names: []string{player.Name},
				Type: VICTORY,
				Received: []*Input{},
			}}
		}
	}

	// If input reqs is empty push to the next player
	if len(r.InputReqs) == 0 {
		p, pidx := r.GetPlayer(r.CurrentPlayer)
		if (p == nil) {
			log.Fatalln("Expected", r.CurrentPlayer, "to exist")
		}
		r.CurrentPlayer = r.Players[(pidx + 1) % len(r.Players)].Name
		r.InputReqs = append(r.InputReqs, &InputRequest{
			Names: []string{r.CurrentPlayer},
			Type: MOVE,
			Received: []*Input{},
		})
	}
	return true, err
}