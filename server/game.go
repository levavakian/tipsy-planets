package main

import (
	"math/rand"
	"sync"
	"time"
	"fmt"
	"log"
	"github.com/gorilla/websocket"
	"errors"
	"github.com/google/uuid"
)

const (
	MOVE = "MOVE"
	BATTLE = "BATTLE"
	VICTORY = "VICTORY"
)

const (
	KNOCKBACK = "KNOCKBACK"
	WORMHOLE = "WORMHOLE"
	GENERIC = "GENERIC"
	TURNSKIP = "TURNSKIP"
)

const (
	EXTERNAL = "EXTERNAL"
	BUILTIN = "BUILTIN"
	ONBATTLELOSE = "ONBATTLELOSE"
	ONBATTLEWIN = "ONBATTLEWIN"
	ONBATTLE = "ONBATTLE"
)

const (
	DICE_SIZE = 6
)

type Prompts struct {
	MaxPriority float64 `json:"max_priority"`
	Priority float64 `json:"priority"`
	PriorityChange float64 `json:"priority_change"`
	Prompts []string `json:"prompts"`
}

type PromptCategory struct {
	MaxPriority float64 `json:"max_priority"`
	Priority float64 `json:"priority"`
	PriorityChange float64 `json:"priority_change"`
	Prompts map[string]*Prompts `json:"prompts"`
}

func newPromptsMapping() map[string]*PromptCategory {
	ret := map[string]*PromptCategory{}
	ret["Truth"] = &PromptCategory{
		Prompts: map[string]*Prompts{
			"Mild": &Prompts{
				MaxPriority: .7,
				Priority: .7,
				PriorityChange: .3,
				Prompts: []string{
					"example mild truth prompt 1",
					"example mild truth prompt 2",
				},
			},
			"Medium": &Prompts{
				MaxPriority: 1.0,
				Priority: .3,
				PriorityChange: .3,
				Prompts: []string{
					"example medium truth prompt 1",
					"example medium truth prompt 2",
				},
			},
			"Spicy": &Prompts{
				MaxPriority: .7,
				Priority: .1,
				PriorityChange: .3,
				Prompts: []string{
					"example spicy truth prompt 1",
					"example spicy truth prompt 2",
				},
			},
		},
	}
	ret["Dare"] = &PromptCategory{
		Prompts: map[string]*Prompts{
			"Mild": &Prompts{
				MaxPriority: .7,
				Priority: .7,
				PriorityChange: .3,
				Prompts: []string{
					"example mild dare prompt 1",
					"example mild dare prompt 2",
				},
			},
			"Medium": &Prompts{
				MaxPriority: 1.0,
				Priority: .3,
				PriorityChange: .3,
				Prompts: []string{
					"example medium dare prompt 1",
					"example medium dare prompt 2",
				},
			},
			"Spicy": &Prompts{
				MaxPriority: .7,
				Priority: .1,
				PriorityChange: .3,
				Prompts: []string{
					"example spicy dare prompt 1",
					"example spicy dare prompt 2",
				},
			},
		},
	}
	ret["Rule"] = &PromptCategory{
		Prompts: map[string]*Prompts{
			"Mild": &Prompts{
				MaxPriority: .7,
				Priority: .7,
				PriorityChange: .3,
				Prompts: []string{
					"example mild rule prompt 1",
					"example mild rule prompt 2",
				},
			},
			"Medium": &Prompts{
				MaxPriority: 1.0,
				Priority: .3,
				PriorityChange: .3,
				Prompts: []string{
					"example medium rule prompt 1",
					"example medium rule prompt 2",
				},
			},
			"Spicy": &Prompts{
				MaxPriority: .7,
				Priority: .1,
				PriorityChange: .3,
				Prompts: []string{
					"example spicy rule prompt 1",
					"example spicy rule prompt 2",
				},
			},
		},
	}
	return ret
}

type Settings struct {
	RequireExactVictory bool `json:"require_exact_victory"`
}

type Player struct {
	Name string `json:"name"`
	Location string `json:"location"`
	Conns map[*websocket.Conn]bool `json:"-"`
}

type LocationEffect struct {
	Id string `json:"id"`
	Type string `json:"type"`
	WormholeTarget string `json:"wormhole_target"`
	KnockbackAmount int `json:"knockback_amount"`
	TurnskipAmount int `json:"turnskip_amount"`
	FlavorText string `json:"flavor_text"`
	Trigger string `json:"trigger"`
}

type Location struct {
	Name string `json:"name"`
	X int `json:"x"`
	Y int `json:"y"`
	Effects []*LocationEffect `json:"effects"`
}

type GameBoard struct {
	Locations []*Location `json:"locations"`
	Effects []*LocationEffect `json:"effects"`
}

func (r *Room) AddEffect(name string, etype string, trigger string, locations []string, flavorText string, knockbackAmount int, wormholeTarget string, turnskipAmount int) {
	g := r.Board
	eff := &LocationEffect{
		Type: etype,
		FlavorText: flavorText,
		KnockbackAmount: knockbackAmount,
		WormholeTarget: wormholeTarget,
		TurnskipAmount: turnskipAmount,
		Trigger: trigger,
		Id: uuid.New().String(),
	}
	if len(locations) == 0 {
		g.Effects = append(g.Effects, eff)
	} else {
		for _, loc := range(g.Locations) {
			add := func()bool {
				for _, target := range locations {
					if target == loc.Name {
						return true
					}
				}
				return false
			}()
			if add {
				loc.Effects = append(loc.Effects, eff)
			}
		}
	}

	// Lets check for the victory condition
	if (len(r.InputReqs) <= 0) {
		return
	}
	ireq := r.InputReqs[0]
	if ireq.Type != VICTORY {
		return
	}
	if (len(ireq.Names) <= 0) {
		return
	}
	if (ireq.Names[0] != name) {
		return
	}
	// Eyy, somebody won clear the board and update the spiciness ratios
	r.PopInputReq()

	for _, cat := range r.Prompts {
		cat.Priority = cat.Priority + (cat.MaxPriority - cat.Priority) * cat.PriorityChange
		if cat.Priority > cat.MaxPriority {
			cat.Priority = cat.MaxPriority
		}
		for _, prompts := range cat.Prompts {
			prompts.Priority = prompts.Priority + (prompts.MaxPriority - prompts.Priority) * prompts.PriorityChange
			if prompts.Priority > prompts.MaxPriority {
				prompts.Priority = prompts.MaxPriority
			}
		}
	}
}

func (r *Room) RemoveEffect(id string) {
	g := r.Board
	nGlobalEffects := []*LocationEffect{}
	for _, eff := range g.Effects {
		if eff.Id != id {
			nGlobalEffects = append(nGlobalEffects, eff)
		}
	}
	g.Effects = nGlobalEffects

	for _, loc := range g.Locations {
		nLocEffects := []*LocationEffect{}
		for _, eff := range loc.Effects {
			if eff.Id != id {
				nLocEffects = append(nLocEffects, eff)
			}
		}
		loc.Effects = nLocEffects
	}
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
	Board *GameBoard `json:"board"`
	LastUpdate time.Time `json:"last_update"`
	InputReqs []*InputRequest `json:"input_reqs"`
	History []string `json:"history"`
	Settings Settings `json:"settings"`
	TurnSkips map[string]int `json:"turn_skips"`
	Prompts map[string]*PromptCategory `json:"prompts"`
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
		TurnSkips: map[string]int{},
		Prompts: newPromptsMapping(),
	}
}

func (r *Room) PopInputReq() {
	r.InputReqs[0] = nil
	r.InputReqs = r.InputReqs[1:]
}

func (r *Room) ClearPendingForPlayer(name string) {
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

func (r *Room) PendingForPlayer(name string, rtype string) bool {
	for _, req := range r.InputReqs {
		if rtype != "" && req.Type != rtype {
			continue
		}
		for _, n := range req.Names {
			if n == name {
				return true
			}
		}
	}
	return false
}

func defaultGameBoard() *GameBoard {
	locs := []*Location{
		{"[1]Start", 183, 420, []*LocationEffect{}},
		{"[2]", 105, 380, []*LocationEffect{}},
		{"[3]", 103, 299, []*LocationEffect{}},
		{"[4]", 163, 252, []*LocationEffect{}},
		{"[5]Spider Hole", 224, 275, []*LocationEffect{{Type: KNOCKBACK, KnockbackAmount: 1, FlavorText: "The spiders scare %s back! They take a drink to settle their nerves."}}},
		{"[6]", 265, 362, []*LocationEffect{}},
		{"[7]Asteroids", 341, 392, []*LocationEffect{{Type: GENERIC, FlavorText: "Asteroids knock a drink into %s's mouth!"}}},
		{"[8]", 393, 456, []*LocationEffect{}},
		{"[9]", 411, 551, []*LocationEffect{}},
		{"[10]Wormhole Chi-Alpha", 427, 617, []*LocationEffect{{Type: WORMHOLE, WormholeTarget: "[14]Wormhole Chi-Beta", FlavorText: "The wormhole sucks %s to Wormhole Chi-Beta and a drink into their mouth!"}}},
		{"[11]", 488, 684, []*LocationEffect{}},
		{"[12]Spacewhale Harbor", 568, 694, []*LocationEffect{{Type: TURNSKIP, TurnskipAmount: 1, FlavorText: "%s is entranced by space whales, they skip a turn!"}}},
		{"[13]", 621, 621, []*LocationEffect{}},
		{"[14]Wormhole Chi-Beta", 561, 543, []*LocationEffect{{Type: WORMHOLE, WormholeTarget: "[10]Wormhole Chi-Alpha", FlavorText: "The wormhole sucks %s to Wormhole Chi-Alpha and two drinks into their mouth!"}}},
		{"[15]", 503, 486, []*LocationEffect{}},
		{"[16]", 487, 409, []*LocationEffect{}},
		{"[17]", 565, 347, []*LocationEffect{}},
		{"[18]Wormhole Tau-Epsilon", 587, 267, []*LocationEffect{{Type: WORMHOLE, WormholeTarget: "[23]Wormhole Tau-Gamma", FlavorText: "The wormhole sucks %s to Wormhole Tau-Gamma and a drink into their mouth!"}}},
		{"[19]", 533, 219, []*LocationEffect{}},
		{"[20]Asteroids", 497, 145, []*LocationEffect{{Type: GENERIC, FlavorText: "Asteroids knock a drink into %s's mouth!"}}},
		{"[21]", 563, 108, []*LocationEffect{}},
		{"[22]", 621, 162, []*LocationEffect{}},
		{"[23]Wormhole Tau-Gamma", 677, 219, []*LocationEffect{{Type: WORMHOLE, WormholeTarget: "[18]Wormhole Tau-Epsilon", FlavorText: "The wormhole sucks %s to Wormhole Tau-Epsilon and a drink into their mouth!"}}},
		{"[24]", 727, 273, []*LocationEffect{}},
		{"[25]", 678, 362, []*LocationEffect{}},
		{"[26]The Spider House", 680, 438, []*LocationEffect{{Type: KNOCKBACK, KnockbackAmount: 2, FlavorText: "The spiders drag %s back! They take two drinks to settle their nerves."}}},
		{"[27]", 714, 517, []*LocationEffect{}},
		{"[28]", 789, 538, []*LocationEffect{}},
		{"[29]", 880, 517, []*LocationEffect{}},
		{"[30]", 913, 437, []*LocationEffect{}},
		{"[31]", 851, 393, []*LocationEffect{}},
		{"[32]Tentomon's Trove", 790, 416, []*LocationEffect{{Type: GENERIC, FlavorText: "A vicious space octopus uses all its tentacles to make %s drink eight times!"}}},
		{"[33]", 761, 487, []*LocationEffect{}},
		{"[34]", 775, 573, []*LocationEffect{}},
		{"[35]Asteroids", 819, 626, []*LocationEffect{{Type: GENERIC, FlavorText: "Asteroids knock a drink into %s's mouth!"}}},
		{"[36]", 895, 639, []*LocationEffect{}},
		{"[37]Solar Storm", 964, 598, []*LocationEffect{{Type: KNOCKBACK, KnockbackAmount: 3, FlavorText: "Solar squalls push %s back! The only cure to the radiation poisoning is to take three drinks."}}},
		{"[38]", 1000, 525, []*LocationEffect{}},
		{"[39]", 1017, 435, []*LocationEffect{}},
		{"[40]Solar Sail", 1019, 358, []*LocationEffect{{Type: KNOCKBACK, KnockbackAmount: 2, FlavorText: "Solar winds push %s back! Drink two to refill your sails."}}},
		{"[41]", 945, 304, []*LocationEffect{}},
		{"[42]Baby Tentomon", 880, 247, []*LocationEffect{{Type: GENERIC, FlavorText: "The space octopus child! It only has four arms to make %s drink four times"}}},
		{"[43]", 907, 176, []*LocationEffect{}},
		{"[44]The Restaurant at the End of the Universe", 1024, 108, []*LocationEffect{{Type: GENERIC, FlavorText: "%s made it! Have a drink and make a new rule."}}},
	}
	for _, loc := range locs {
		for _, eff := range loc.Effects {
			eff.Id = uuid.New().String()
			eff.Trigger = BUILTIN
		}
	}
	return &GameBoard{
		Effects: []*LocationEffect{},
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
	err := r.MovePlayer(input.Received[0].Name, dice, []string{}, false)
	if err != nil {
		return err
	}

	r.PopInputReq()
	return nil
}

func (r *Room) DoVictory(input *InputRequest) error {
	// We wait for a new rule so do nothing except clear the received
	input.Received = []*Input{}
	return nil
}

func (r *Room)MovePlayer(name string, amount int, prevLocsThisRound []string, forced bool) error {
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

	// If no battles are pending for the player, do location effects
	if !r.PendingForPlayer(player.Name, BATTLE) {
		prevLocsThisRound = append(prevLocsThisRound, player.Location)
		err := r.DoEffects(player, EXTERNAL, prevLocsThisRound, false)
		err = r.DoEffects(player, BUILTIN, prevLocsThisRound, false)
		if err != nil {
			return err
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


	winner, loser, diff := func()(*Player, *Player, int){
		if rollOne > rollTwo {
			return playerOne, playerTwo, rollTwo - rollOne
		} else {
			return playerTwo, playerOne, rollOne - rollTwo
		}
	}()

	err := r.DoEffects(playerOne, ONBATTLE, []string{winner.Location}, true)
	if err != nil {
		return err
	}
	err = r.DoEffects(playerTwo, ONBATTLE, []string{winner.Location}, true)
	if err != nil {
		return err
	}

	if diff == 0 {
		r.History = append(r.History, "Battle was a tie!")
		r.PopInputReq()
		r.InputReqs = append([]*InputRequest{&InputRequest{
			Names: []string{playerOne.Name, playerTwo.Name},
			Type: BATTLE,
			Received: []*Input{},
		}}, r.InputReqs...)
		return nil
	} else {
		err = r.DoEffects(winner, ONBATTLEWIN, []string{winner.Location}, true)
		if err != nil {
			return err
		}
		err = r.DoEffects(loser, ONBATTLELOSE, []string{winner.Location}, true)
		if err != nil {
			return err
		}
	}

	r.ClearPendingForPlayer(loser.Name)
	r.PopInputReq()

	r.MovePlayer(loser.Name, diff, []string{}, true)
	if winner.Name == playerOne.Name {
		// If there's no more battles for playerOne, apply effects
 		if !r.PendingForPlayer(winner.Name, BATTLE) {
			err = r.DoEffects(playerOne, EXTERNAL, []string{winner.Location}, false)
			err = r.DoEffects(playerOne, BUILTIN, []string{winner.Location}, false)
			if err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (r *Room) DoEffects(p *Player, triggerType string, prevLocsThisRound []string, generic bool) error {
	location, lidx := r.Board.GetLocation(p.Location)
	if location == nil {
		return errors.New(p.Location + " did not exist")
	}

	haveVisited := func(visitTarget string)bool{
		for _, n := range prevLocsThisRound {
			if n == visitTarget {
				return true
			}
		}
		return false
	}

	deferred_move_diff := 0

	targetList := location.Effects
	if generic {
		targetList = append(targetList, r.Board.Effects...)
	}
	for _, effect := range targetList {
		if effect.Trigger != triggerType {
			continue
		}
		switch effect.Type {
		case WORMHOLE:
			if deferred_move_diff != 0 {
				continue
			}
			target, tidx := r.Board.GetLocation(effect.WormholeTarget)
			if target == nil {
				return errors.New(effect.WormholeTarget + "did not exist for wormhole")
			}
			if haveVisited(effect.WormholeTarget) {
				continue
			}
			diff := tidx - lidx
			r.History = append(r.History, fmt.Sprintf(effect.FlavorText, p.Name))
			deferred_move_diff = diff
		case KNOCKBACK:
			if deferred_move_diff != 0 {
				continue
			}
			tidx := lidx - effect.KnockbackAmount
			lastIdx := len(r.Board.Locations) - 1
			if tidx > lastIdx {
				tidx = lastIdx - (tidx - lastIdx)
			}
			if tidx < 0 {
				tidx = 0
			}
			diff := tidx - lidx
			target := r.Board.Locations[tidx]
			if haveVisited(target.Name) {
				continue
			}
			r.History = append(r.History, fmt.Sprintf(effect.FlavorText, p.Name))
			deferred_move_diff = diff
		case TURNSKIP:
			r.TurnSkips[p.Name] = r.TurnSkips[p.Name] + effect.TurnskipAmount
			r.History = append(r.History, fmt.Sprintf(effect.FlavorText, p.Name))
		case GENERIC:
			r.History = append(r.History, fmt.Sprintf(effect.FlavorText, p.Name))
		default:
			return errors.New("Hit default case in effects switch")
		}
	}
	if deferred_move_diff == 0 {
		return nil
	} else {
		return r.MovePlayer(p.Name, deferred_move_diff, prevLocsThisRound, true)
	}
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
	default:
		return true, errors.New("Hit default case in input request switch")
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
		for {
			p, pidx := r.GetPlayer(r.CurrentPlayer)
			if (p == nil) {
				log.Fatalln("Expected", r.CurrentPlayer, "to exist")
			}
			r.CurrentPlayer = r.Players[(pidx + 1) % len(r.Players)].Name
			if r.TurnSkips[r.CurrentPlayer] > 0 {
				r.TurnSkips[r.CurrentPlayer] = r.TurnSkips[r.CurrentPlayer] - 1
				r.History = append(r.History, fmt.Sprintf("Skipped %s's turn", r.CurrentPlayer))
				continue
			}
			r.InputReqs = append(r.InputReqs, &InputRequest{
				Names: []string{r.CurrentPlayer},
				Type: MOVE,
				Received: []*Input{},
			})
			break
		}
	}
	return true, err
}