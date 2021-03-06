import * as paper from "paper";

enum EffectTypes {
  WORMHOLE = "WORMHOLE",
  KNOCKBACK = "KNOCKBACK",
  GENERIC = "GENERIC",
  TURNSKIP = "TURNSKIP",
}

enum TriggerTypes {
  EXTERNAL = "EXTERNAL",
  ONBATTLELOSE = "ONBATTLELOSE",
	ONBATTLEWIN = "ONBATTLEWIN",
  ONBATTLE = "ONBATTLE",
  BUILTIN = "BUILTIN",
}

enum InputTypes {
  BATTLE = "BATTLE",
  MOVE = "MOVE",
  VICTORY = "VICTORY",
}

class Prompts {
  max_priority: number;
  priority: number;
  priority_change: number;
  prompts: string[];

  constructor(props: any) {
    this.max_priority = props.max_priority
    this.priority = props.priority
    this.priority_change = props.priority_change
    this.prompts = props.prompts
  }
}

class PromptCategory {
  max_priority: number;
  priority: number;
  priority_change: number;
  prompts: Map<string, Prompts>;

  constructor(props: any) {
    this.max_priority = props.max_priority
    this.priority = props.priority
    this.priority_change = props.priority_change
    this.prompts = new Map<string, Prompts>()
    for (let key in props.prompts) {
      this.prompts.set(key, new Prompts(props.prompts[key]))
    }
  }
}

class Player {
  name: string;
  location: string;

  constructor(props: any) {
    this.name = props.name
    this.location = props.location
  }
}

class LocationEffect {
  id: string
  type: string
  wormhole_target: string
  knockback_amount: number
  turnskip_amount: number
  flavor_text: string
  trigger: string

  constructor(props: any) {
    this.id = props.id
    this.type = props.type
    this.wormhole_target = props.wormhole_target
    this.knockback_amount = props.knockback_amount
    this.turnskip_amount = props.turnskip_amount
    this.flavor_text = props.flavor_text
    this.trigger = props.trigger
  }
}

class Location {
  name: string
  x: number
  y: number
  effects: LocationEffect[]

  constructor(props: any) {
    this.name = props.name
    this.x = props.x
    this.y = props.y
    this.effects = []
    for (let effprop of props.effects) {
      this.effects.push(new LocationEffect(effprop))
    }
  }
}

class GameBoard {
  locations: Location[];
  effects: LocationEffect[]

  constructor(props: any) {
    this.locations = []
    for (let locjson of props.locations) {
      this.locations.push(new Location(locjson))
    }
    this.effects = []
    for (let effprop of props.effects) {
      this.effects.push(new LocationEffect(effprop))
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
  prompts: Map<string, PromptCategory>

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
    this.prompts = new Map<string, PromptCategory>()
    for (let key in props.prompts) {
      this.prompts.set(key, new PromptCategory(props.prompts[key]))
    }
  }
}

const getPlayerColor = (idx: number) => {
  let color = new paper.Color("red")
  color.hue += idx*230
  return color
}

export { Room, Player, GameBoard, InputRequest, Location, LocationEffect, TriggerTypes, EffectTypes, InputTypes, getPlayerColor, Prompts, PromptCategory }