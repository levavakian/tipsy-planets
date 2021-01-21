import React from 'react';
import { toast } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import { api } from './api'
import { Room, EffectTypes, TriggerTypes, Location, LocationEffect } from './Elements'

interface RulesProps {
  room?: Room
  name: string
  lobby: string
}

interface RulesState {
  type: string
  flavor_text: string
  locations: string[]
  location: string
}

class Rules extends React.Component<RulesProps,RulesState> {
  id: number

  constructor(props: RulesProps) {
    super(props)
    this.state = {
      type: TriggerTypes.EXTERNAL,
      flavor_text: "",
      locations: [],
      location: ""
    }
    this.id = 0
  }

  newSeq = () => {
    this.id = this.id + 1
    return this.id
  }

  onTextChange = (ev: any) => {
    ev.preventDefault()
    this.setState({
      flavor_text: ev.target.value
    })
  }

  onTypeChange = (ev: any) => {
    ev.preventDefault()
    this.setState({
      type: ev.target.value
    })
  }

  onLocationChange = (ev: any) => {
    ev.preventDefault()
    this.setState({
      location: ev.target.value
    })
  }

  onLocationSubmit = (ev: any) => {
    ev.preventDefault()
    this.setState((prevState) => {
      let new_locs = [...prevState.locations]
      new_locs.push(this.state.location)
      return {
        locations: new_locs
      }
    })
  }

  onClear = (ev: any) => {
    ev.preventDefault()
    this.setState({
        type: TriggerTypes.EXTERNAL,
        flavor_text: "",
        locations: [],
        location: ""
    })
  }

  makeRule() {
    return (
      <div className="Flexcolumn">
        <span>Rule</span>
        <div className="Flexrow">
          <select value={this.state.type} onChange={this.onTypeChange} id="type">
            {Object.keys(TriggerTypes).filter((val: string) => {return val !== TriggerTypes.BUILTIN}).map((val: string) => {
              return <option value={val}>{val}</option>
            })}
          </select>
          <span>&nbsp;&nbsp;</span>
          <select value={this.state.location} id="locations" onChange={this.onLocationChange}>
            <option value="">Location</option>
            {this.props.room?.board.locations.map((loc) => {
              return <option key={loc.name} value={loc.name}>{loc.name}</option>
            })}
          </select>
          <div className="Flexcolumn">
            <span onClick={this.onLocationSubmit} className="cardanim buttonlist">add loc</span>
            <div className="Flexrow">
              {this.state.locations.map((loc) => {return <span key={this.newSeq()}>{loc}&nbsp;</span>})}
            </div>
          </div>
          <span>&nbsp;&nbsp;</span>
        </div>
        <input style={{width: "100%", textAlign: "left"}} value={this.state.flavor_text} onChange={this.onTextChange} placeholder="rule text"></input>
        <div className="Flexrow">
          <span onClick={this.clearRule} className="cardanim buttonlist">Clear</span>
          <span onClick={this.submitRule} className="cardanim buttonlist">Submit</span>
        </div>
      </div>
    )
  }

  clearRule = (ev: any) => {
    this.setState({
      locations: [],
      location: "",
      type: TriggerTypes.EXTERNAL,
      flavor_text: "",
    })
  }

  deleteRule = (ev: any, eff: LocationEffect) => {
    ev.preventDefault()
    if (!this.props.room) {
      return
    }
    api("POST", "rule", {"code": this.props.room.code, "delete": true, "id": eff.id}, (e: any) => {
      if (e.target.status !== 200) {
        toast(e.target.response?.error)
        return
      }
    })
  }

  submitRule = (ev: any) => {
    ev.preventDefault()
    if (!this.props.room) {
      return
    }
    let ftext = this.state.flavor_text
    if (!ftext.includes("%s")) {
      ftext = "%s: " + ftext
    }
    let content = {
      "type": EffectTypes.GENERIC,
      "delete": false,
      "trigger": this.state.type,
      "locations": this.state.locations,
      "code": this.props.room.code,
      "flavor_text": ftext,
      "name": this.props.name,
    }
    api("POST", "rule", content, (e: any) => {
      if (e.target.status !== 200) {
        toast(e.target.response?.error)
        return
      }
    })
    this.clearRule(ev)
  }

  getRuleString = (effect: LocationEffect) => {
    let re = "%s"
    let s = ""
    if (effect.trigger !== "") {
      s += `[${effect.trigger}]`
    }
    s += `${effect.flavor_text.replace(re, "PLAYER")}`
    return s
  }

  displayRules = () => {
    let global_rules = (
      this.props.room?.board.effects.filter((e: LocationEffect) => { return e.trigger !== TriggerTypes.BUILTIN }).map((e: LocationEffect) => {
        return (
          <div key={e.id} className="Flexcolumn">
            <span key={e.id}>{this.getRuleString(e)}</span>
            <span className="buttonlist cardanim" key={e.id+"delete"} onClick={(ev: any) => this.deleteRule(ev, e)}>DELETE</span>
          </div>
        )
      })
    )
    let location_rules = (
      this.props.room?.board.locations.map((l: Location) => {
        if (l.effects.filter((e: LocationEffect) => { return e.trigger !== TriggerTypes.BUILTIN }).length === 0) {
          return <></>
        } 
        return (
          <div key={l.name} className="card buttonlist">
            {l.name}
            {l.effects.filter((e: LocationEffect) => { return e.trigger !== TriggerTypes.BUILTIN }).map((e: LocationEffect) => {
              return (
                <div key={l.name + e.id} className="Flexcolumn">
                  <span key={e.id}>{this.getRuleString(e)}</span>
                  <span className="buttonlist cardanim" key={e.id+"delete"} onClick={(ev: any) => this.deleteRule(ev, e)}>DELETE</span>
                </div>
              )
            })}
          </div>
        )
      })
    )
    return (
      <div className="Flexcolumn">
        <div className="buttonlist">
          Global rules
          {global_rules}
        </div>
        <div className="buttonlist">
          Location rules
          {location_rules}
        </div>
      </div>
    )
  }

  render() {
    return (
      <div>
        <div className="buttonlist Flexrow">
          {this.makeRule()}
        </div>
        <div className="card buttonlist">
          {this.displayRules()}
        </div>
      </div>
      
    )
  }
}

export default Rules;