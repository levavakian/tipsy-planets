import React from 'react';
import { ToastContainer, toast } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import { api, wsURL } from './api'
import { Room } from './Elements'
import Interaction from './Interaction'
import History from './History'
import Canvas from './Canvas'
import Prompts from './Prompts'

interface LobbyProps {
  lobby: string;
  name: string;
}

interface LobbyState {
  room?: Room
  img?: paper.Raster
}
  
class Lobby extends React.Component<LobbyProps, LobbyState> {
  timerId?: number
  ws?: WebSocket
  last_ws_update: Date

  constructor(props: LobbyProps) {
    super(props)
    this.last_ws_update = new Date()
    this.state = {
    }
  }

  componentDidMount() {
    this.loadFromServer()
    this.ws = this.makeWS()
    this.timerId = window.setInterval(
      () => this.poll(),
      10000
    )
  }

  makeWS() {
    let socket = new WebSocket(wsURL + `/api/stream?name=${this.props.name}&code=${this.props.lobby}`)
    socket.onmessage = (ev: MessageEvent<any>) => {
      this.last_ws_update = new Date()
      let parsed = JSON.parse(ev.data)
      if ('heartbeat' in parsed) {
        return
      }
      if ('ping' in parsed) {
        toast(parsed.ping + " asks that you hurry up")
        return
      }
      this.loadFromServer()
    }
    return socket
  }

  poll() {
    let now = new Date()
    let timeDiff = (now.getTime() - this.last_ws_update.getTime()) / 1000
    if (timeDiff > 10)
    {
      this.loadFromServer()
      this.ws?.close()
      this.ws = this.makeWS()
    }
  }

  componentWillUnmount() {
    if (this.timerId) {
      clearInterval(this.timerId)
    }
  }

  loadFromServer() {
    api("POST", "state", {"code": this.props.lobby}, (e: any) => {
      if (e.target.status !== 200) {
        toast("error", e.target.response?.error)
        return
      }
      let room = new Room(e.target.response)
      this.setState((prevState) => {
        return {
          room: room
        }
      })
    })
  }

  render() {
    return (
      <div>
        <ToastContainer />
        <div className="App-banner">Tipsy Planets || Lobby is {this.props.lobby} || Name is {this.props.name}
          { this.state.room ? <Canvas room={this.state.room} /> : <></> }
          <div className="card">
            <Interaction
              lobby={this.props.lobby}
              name={this.props.name}
              room={this.state.room} />
          </div>
          <History room={this.state.room} />
          <Prompts room={this.state.room} />
        </div>
      </div>
    )
  }
}

export default Lobby;