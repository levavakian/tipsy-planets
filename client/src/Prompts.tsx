import React from 'react';
import { Room, PromptCategory } from './Elements'
import { api } from './api'
import { toast } from 'react-toastify';

interface PromptsProps {
  room?: Room
}

interface PromptsState {
  latest_prompt: string
}

class Prompts extends React.Component<PromptsProps,PromptsState> {
  constructor(props: PromptsProps) {
    super(props)
    this.state = {
      latest_prompt: ""
    }
  }
  requestPrompt = (category: string, level: string) => {
    if (!this.props.room) {
      return
    }
    api("POST", "prompt", {"code": this.props.room.code, "category": category, "level": level}, (e: any) => {
      if (e.target.status !== 200) {
        toast(e.target.response?.error)
        return
      }
      this.setState({
        latest_prompt: e.target.response?.prompt
      })
    })
  }

  promptCategory = (key: string, props: PromptCategory) => {
    let levelJSX = Array.from(props.prompts).map((val) => {
      let lkey = val[0]
      return <div onClick={(evt: any) => {this.requestPrompt(key, lkey)}} className="cardanim buttonlist" key={lkey}>{lkey}</div>
    })
    return (
      <div>
        <div onClick={(evt: any) => {this.requestPrompt(key, "")}} className="cardanim buttonlist">{key}</div>
        {levelJSX}
      </div>
    )
  }

  render() {
    if (!this.props.room) {
      return <div></div>
    }

    let catJSX = Array.from(this.props.room.prompts.entries()).map((val) => {
      let [key, cat] = val
      return this.promptCategory(key, cat)
    })

    return (
      <div>
        <span>Latest prompt: {this.state.latest_prompt}</span>
        <div className="Flexrow">
          {catJSX}
        </div>
      </div>
      
    )
  }
}

export default Prompts