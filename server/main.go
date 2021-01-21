package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
	"encoding/json"
	"strconv"
	"github.com/markbates/pkger"
	"github.com/gorilla/websocket"
	"os"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func RandStringRunes(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

func setupHeaders(w *http.ResponseWriter, req *http.Request) bool {
	if nocors := os.Getenv("NOCORS"); nocors != "" {
		(*w).Header().Set("Access-Control-Allow-Origin", "*")
	}
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")
	(*w).Header().Set("Content-Type", "application/json")

	if req.Method != http.MethodPost && req.Method != http.MethodGet {
		(*w).WriteHeader(http.StatusOK)
		return false
	}
	return true
 }

type JSONError struct {
	Error string `json:"error"`
}

func WriteError(w http.ResponseWriter ,err string, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(JSONError{err})
}

func (r *Room) NotifyPlayers() {
	for _, player := range r.Players {
		for ws, _ := range player.Conns {
			err := ws.WriteJSON(struct{}{})
			if err != nil {
				ws.Close()
				delete(player.Conns, ws)
			}
		}
	}
}

type LockedRooms struct {
	sync.RWMutex
	Rooms map[string]*Room
}

func HandleCreate(rooms *LockedRooms) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !setupHeaders(&w, r) {
			return
		}

		rooms.Lock()
		defer rooms.Unlock()

		type CreateRes struct {
			Code string `json:"code"`
		}
		
		for i := 0; i < 10000; i++ {
			code := &CreateRes{Code: RandStringRunes(6)}

			if _, ok := rooms.Rooms[code.Code]; ok {
				continue
			}

			rooms.Rooms[code.Code] = newRoom(code.Code)
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(code)
			return
		}

		WriteError(w, "could not create unique room code", http.StatusInternalServerError)
		return
	}
}

func HandleJoin(rooms *LockedRooms) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !setupHeaders(&w, r) {
			return
		}

		type JoinReq struct {
			Code string
			Name string
		}
		var joinReq JoinReq
		err := json.NewDecoder(r.Body).Decode(&joinReq)
		if err != nil {
			WriteError(w, err.Error(), http.StatusBadRequest)
			return
		}
		if joinReq.Code == "" || joinReq.Name == "" {
			WriteError(w, "name or lobby code missing from join request", http.StatusBadRequest)
			return
		}

		rooms.Lock()
		room, ok := rooms.Rooms[joinReq.Code]
		rooms.Unlock()

		if !ok {
			WriteError(w, "tried to join nonexistant lobby", http.StatusBadRequest)
			return
		}

		room.Lock()
		defer room.Unlock()

		for _, player := range room.Players {
			if player.Name == joinReq.Name {
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(room)
				return
			}
		}

		newPlayer := &Player{Name: joinReq.Name, Conns: map[*websocket.Conn]bool{}, Location: room.Board.Locations[0].Name}
		room.Players = append(room.Players, newPlayer)
		room.LastUpdate = time.Now()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(room)
		room.NotifyPlayers()

		return
	}
}

func HandleBoardState(rooms *LockedRooms) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !setupHeaders(&w, r) {
			return
		}
		
		type StateReq struct {
			Code string
		}
		var stateReq StateReq
		err := json.NewDecoder(r.Body).Decode(&stateReq)
		if err != nil {
			WriteError(w, err.Error(), http.StatusBadRequest)
			return
		}
		if stateReq.Code == "" {
			WriteError(w, "lobby code missing from join request", http.StatusBadRequest)
			return
		}

		rooms.Lock()
		room, ok := rooms.Rooms[stateReq.Code]
		rooms.Unlock()

		if !ok {
			WriteError(w, "tried to get board state for nonexistant lobby", http.StatusBadRequest)
			return
		}

		room.Lock()
		defer room.Unlock()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(room)
		return
	}
}

func HandleImage(img []byte) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Length", strconv.Itoa(len(img)))

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		w.WriteHeader(http.StatusAccepted)
		_, err := w.Write(img)
		if err != nil {
			WriteError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		return
	}
}

func getImage() []byte {
	imgf, err := pkger.Open("/gameboard.jpg")
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer imgf.Close()

	info, err := imgf.Stat()
	if err != nil {
		log.Fatalln(err.Error())
	}

	imgbytes := make([]byte, info.Size())
	_, err = imgf.Read(imgbytes)
	if err != nil {
		log.Fatalln(err)
	}

	return imgbytes
}

func HandleStream(rooms *LockedRooms, upgrader *websocket.Upgrader) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		codes, ok := r.URL.Query()["code"]
		if !ok || len(codes) == 0{
			WriteError(w, "did not have room code in request", http.StatusBadRequest)
			return
		}
		code := codes[0]

		names, ok := r.URL.Query()["name"]
		if !ok || len(names) == 0 {
			WriteError(w, "did not have player name in request", http.StatusBadRequest)
			return
		}
		name := names[0]

		rooms.Lock()
		room, ok := rooms.Rooms[code]
		rooms.Unlock()

		if !ok {
			WriteError(w, "tried to start stream for nonexistant lobby", http.StatusBadRequest)
			return
		}

		room.Lock()
		defer room.Unlock()

		type Heartbeat struct {
			Heartbeat bool `json:"heartbeat"`
		}

		for _, player := range room.Players {
			if player.Name == name {
				ws, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					log.Fatalln(err.Error())
				}
				player.Conns[ws] = true

				go func() {
					ticker := time.NewTicker(500 * time.Millisecond)
					for {
						select {
						case  <-ticker.C:
							hb := Heartbeat{}
							errws := ws.WriteJSON(hb)
							if errws != nil {
								return
							}
						}
					}
				}()
				
				return
			}
		}
		WriteError(w, "tried to start stream for nonexistant player", http.StatusBadRequest)
	}
}

func HandleInput(rooms *LockedRooms) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !setupHeaders(&w, r) {
			return
		}

		var input Input
		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			WriteError(w, err.Error(), http.StatusBadRequest)
			return
		}
		if input.Code == "" {
			WriteError(w, "lobby code missing from input", http.StatusBadRequest)
			return
		}
		if input.Name == "" {
			WriteError(w, "name missing from input", http.StatusBadRequest)
			return
		}

		rooms.Lock()
		room, ok := rooms.Rooms[input.Code]
		rooms.Unlock()

		if !ok {
			WriteError(w, "no such lobby", http.StatusBadRequest)
			return
		}

		room.Lock()
		defer room.Unlock()

		changed, err := room.AdvanceRoomState(&input)
		if changed {
			room.NotifyPlayers()
		}
		if err != nil {
			WriteError(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func HandlePrompt(rooms *LockedRooms) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !setupHeaders(&w, r) {
			return
		}

		type PromptReq struct {
			Code string
			Level string
			Category string
		}
		var req PromptReq
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			WriteError(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Code == "" {
			WriteError(w, "lobby code missing from prompt request", http.StatusBadRequest)
			return
		}

		rooms.Lock()
		room, ok := rooms.Rooms[req.Code]
		rooms.Unlock()

		if !ok {
			WriteError(w, "no such lobby", http.StatusBadRequest)
			return
		}

		room.Lock()
		defer room.Unlock()

		cat, ok := room.Prompts[req.Category]
		if !ok {
			WriteError(w, "no such category", http.StatusBadRequest)
		}

		chosen := func()*Prompts {
			if req.Level == "" {
				total := 0.0
				for _, v := range cat.Prompts {
					total = total + v.Priority
				}
				r := rand.Float64() * total

				acc := 0.0
				var last *Prompts
				for _, v := range cat.Prompts {
					last = v
					acc = acc + v.Priority
					if r < acc {
						return v
					}
				}
				return last
			} else {
				level, ok := cat.Prompts[req.Level]
				if !ok {
					return nil
				}
				return level
			}
		}()
		if chosen == nil {
			WriteError(w, "no such level", http.StatusBadRequest)
		}

		type PromptResp struct {
			Prompt string `json:"prompt"`
		}

		resp := PromptResp{
			Prompt: chosen.Prompts[rand.Intn(len(chosen.Prompts))],
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

func HandlePing(rooms *LockedRooms) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !setupHeaders(&w, r) {
			return
		}

		type PingReq struct {
			Code string
			Name string
		}
		var req PingReq
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			WriteError(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Code == "" {
			WriteError(w, "lobby code missing from ping request", http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			WriteError(w, "name missing from ping request", http.StatusBadRequest)
			return
		}

		rooms.Lock()
		room, ok := rooms.Rooms[req.Code]
		rooms.Unlock()

		if !ok {
			WriteError(w, "no such lobby", http.StatusBadRequest)
			return
		}

		room.Lock()
		defer room.Unlock()

		type Ping struct {
			Ping string `json:"ping"`
		}

		if len(room.InputReqs) > 0 {
			recv := map[string]bool{}
			for _, input := range room.InputReqs[0].Received {
				recv[input.Name] = true
			}

			for _, name := range room.InputReqs[0].Names {
				if gotten, _ := recv[name]; !gotten {
					player, _ := room.GetPlayer(name)
					for ws, _ := range player.Conns {
						nerr := ws.WriteJSON(Ping{req.Name})
						if nerr != nil {
							err = nerr
						}
					}
				}
			}
		}

		if err != nil {
			WriteError(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		w.WriteHeader(http.StatusOK)
	}
}

func HandleRule(rooms *LockedRooms) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !setupHeaders(&w, r) {
			return
		}

		type RuleReq struct {
			Code string
			Name string
			Id string
			Delete bool
			Locations []string
			FlavorText string `json:"flavor_text"`
			Type string
			Trigger string
			KnockbackAmount int
			WormholeTarget string
			TurnskipAmount int
		}
		var req RuleReq
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			WriteError(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Code == "" {
			WriteError(w, "lobby code missing from prompt request", http.StatusBadRequest)
			return
		}

		rooms.Lock()
		room, ok := rooms.Rooms[req.Code]
		rooms.Unlock()

		if !ok {
			WriteError(w, "no such lobby", http.StatusBadRequest)
			return
		}

		room.Lock()
		defer room.Unlock()

		if req.Delete {
			room.RemoveEffect(req.Id)
		} else {
			room.AddEffect(req.Name, req.Type, req.Trigger, req.Locations, req.FlavorText,
		        			req.KnockbackAmount, req.WormholeTarget, req.TurnskipAmount)
		}

		w.WriteHeader(http.StatusOK)
		room.NotifyPlayers()
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	host := "0.0.0.0"
	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	rooms := &LockedRooms{Rooms: make(map[string]*Room)}

	checkOrigin := func(r *http.Request)bool{ 
		{ return true }
	}
	upgrader := &websocket.Upgrader{
		CheckOrigin: checkOrigin,
	}

	img := getImage()

	http.HandleFunc("/api/create", HandleCreate(rooms))
	http.HandleFunc("/api/join", HandleJoin(rooms))
	http.HandleFunc("/api/state", HandleBoardState(rooms))
	http.HandleFunc("/api/board", HandleImage(img))
	http.HandleFunc("/api/stream", HandleStream(rooms, upgrader))
	http.HandleFunc("/api/input", HandleInput(rooms))
	http.HandleFunc("/api/prompt", HandlePrompt(rooms))
	http.HandleFunc("/api/ping", HandlePing(rooms))
	http.HandleFunc("/api/rule", HandleRule(rooms))
	http.Handle("/", http.FileServer(http.Dir("/home/apps/tipsy-planets/client/build")))
	log.Println("Game server starting on", host, port)
	log.Println(http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), nil))
}