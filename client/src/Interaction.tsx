import React, { createRef,RefObject } from 'react';
import { toast } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import { api } from './api'
import { Room, InputTypes } from './Elements'

interface InteractionProps {
  room?: Room
  name: string
  lobby: string
}

interface InteractionState {
  hiddenDie: number
}

class Interaction extends React.Component<InteractionProps,InteractionState> {
  canvasRef: RefObject<HTMLCanvasElement>

  constructor(props: InteractionProps) {
    super(props)
    this.canvasRef = createRef<HTMLCanvasElement>()
    this.state = {
      hiddenDie: 0
    }
  }

  onStart = (event: any) => {
    event.preventDefault()
    event.stopPropagation()

    api("POST", "input", {"code": this.props.lobby, "name": this.props.name, "value": 0}, (e: any) => {
      if (e.target.response?.error) {
          toast(e.target.response.error)
      }
    })
  }

  makeMove() {
    if (!this.props.room) {
      return <span>Waiting for room...</span>
    }

    if (this.props.room.input_reqs.length === 0) {
      return <span className="cardanim buttonlist" onClick={this.onStart}>Click to start</span>
    }

    let input_req = this.props.room.input_reqs[0]

    let in_inputs_list = input_req.names.find(e => e === this.props.name)
    if (in_inputs_list) {
      let in_received_list = input_req.received.find(elem => elem.name === this.props.name)
      if (!in_received_list) {
        if (input_req.type === InputTypes.MOVE) {
          return <span className="cardanim buttonlist" onClick={this.onStart}>Make your move</span>
        } else if (input_req.type === InputTypes.VICTORY) {
          return <span className="cardanim buttonlist" onClick={this.onStart}>You won! Make a new rule</span>
        } else if (input_req.type === InputTypes.BATTLE) {
          return <span className="cardanim buttonlist" onClick={this.onStart}>Roll for battle!</span>
        }
      }
    }

    return <span className="cardanim buttonlist">Waiting for: {input_req.names.filter(e => !input_req.received.find(elem => elem.name === e)).join(", ")}</span>
  }

  dieRoll = (evt: any) => {
    this.setState({
      hiddenDie: 1 + Math.floor(Math.random() * Math.floor(6))
    })
  }

  render() {
    return (
      <div className="Flexrow">
        {this.makeMove()}
        <span onClick={this.dieRoll} className="cardanim buttonlist">Hidden Die: {this.state.hiddenDie}</span>
      </div>
    )
  }
}

export default Interaction;