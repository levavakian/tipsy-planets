import * as paper from "paper";

class Player {
  name: string;
  location: string;

  constructor(props: any) {
    this.name = props.name
    this.location = props.location
  }
}

class Location {
  name: string
  x: number
  y: number

  constructor(props: any) {
    this.name = props.name
    this.x = props.x
    this.y = props.y
  }
}

class GameBoard {
  locations: Location[];

  constructor(props: any) {
    this.locations = []
    for (let locjson of props.locations) {
      this.locations.push(new Location(locjson))
    }
  }
}

class Input {
  name: string
  val: number
  code: string

  constructor(props: any) {
    this.name = props.name
    this.val = props.val
    this.code = props.code
  }
}

class InputRequest {
  names: string[]
  type: string
  received: Input[]

  constructor(props: any) {
    this.names = props.names
    this.type = props.type
    this.received = []
    for (let input of props.received) {
      this.received.push(new Input(input))
    }
  }
}

class Room {
  code: string;
  board: GameBoard;
  players: Player[];
  last_update: Date;
  input_reqs: InputRequest[]
  history: string[]

  constructor(props: any) {
    this.code = props.code
    this.board = new GameBoard(props.board)
    this.players = []
    this.history = props.history
    for (let jsonplayer of props.players) {
      this.players.push(new Player(jsonplayer))
    }
    this.last_update = new Date(props.last_update)
    this.input_reqs = []
    for (let req of props.input_reqs) {
      this.input_reqs.push(new InputRequest(req))
    }
  }
}

const getPlayerColor = (idx: number) => {
  let color = new paper.Color("red")
  color.hue += idx*230
  return color
}

export { Room, Player, GameBoard, InputRequest, getPlayerColor }