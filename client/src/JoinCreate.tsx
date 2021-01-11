import { ToastContainer, toast } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import React from 'react';
import { api } from './api'

interface JoinCreateProps {
  switchLobby: (code: string, name: string) => void
}
  
interface JoinCreateState {
  name: string
  join: string
}
  
class JoinCreate extends React.Component<JoinCreateProps, JoinCreateState> {
  constructor(props: JoinCreateProps) {
    super(props)
    this.state = {
      name: "",
      join: ""
  }
}

onNameChange = (event: any) => {
  this.setState((prevState) => {
    return {
      name: event.target.value
    }
  })
}

onJoinChange = (event: any) => {
  this.setState((prevState) => {
    return {
      join: event.target.value
    }
  })
}

onCreate = (event: any) => {
  event.preventDefault()
  event.stopPropagation()
  if (!this.state.name) {
    toast("Set your name before creating lobby")
    return
  }
  api("POST", "create", undefined, (e: any) => {
    if (e.target.status !== 201) {
      toast(e.target.response.error)
      return
    }
    const code = e.target.response.code
    const name = this.state.name
    api("POST", "join", {"code": code, "name": name}, (e: any) => {
      if (e.target.status !== 201) {
          toast(e.target.response.error)
          return
      }
      this.props.switchLobby(code, name)
    })
  })
}

onJoin = (event: any) => {
  if (event.which !== 13) {
    return
  }
  event.preventDefault();
  if (!this.state.name) {
    toast("Set your name before joining lobby")
    return
  }
  if (!this.state.join) {
    toast("Set lobby code before joining lobby")
    return
  }
  const code = this.state.join
  const name = this.state.name
  api("POST", "join", {"code": code, "name": name}, (e: any) => {
    if (e.target.status !== 201) {
      toast(e.target.response?.error)
      return
    }
    this.props.switchLobby(code, name)
  })
}

boop = (event: any) => {
  if (event.which === 13) {
    event.preventDefault();
    toast(event.target.value)
  }
}

render() {
  return (
    <div className="App">
      <ToastContainer />
      <header className="App-header">
        <div>Tipsy Planets</div>
        <div>
          <div className="Flexrow">
            <span className="cardanim buttonlist">Name</span>
            <input value={this.state.name} onChange={this.onNameChange} onKeyPress={this.boop} placeholder="your name"></input>
          </div>
          
          <div className="Flexrow">
            <span className="cardanim buttonlist">Join</span>
            <input value={this.state.join} onChange={this.onJoinChange} onKeyPress={this.onJoin} placeholder="room code"></input>
          </div>
          <div onClick={this.onCreate} className="cardanim buttonlist">Create</div>
        </div>
      </header>
    </div>
  )
}
}

export default JoinCreate;