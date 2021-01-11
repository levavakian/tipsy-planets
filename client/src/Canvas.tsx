import { serverURL } from './api'
import { Room, getPlayerColor } from './Elements'
import * as paper from "paper";
import { Path, Point } from "paper";
import React, { createRef,RefObject } from 'react';
import { Player } from './Elements'

function Playerlist(props: any) {
  return (
    <div className="Flexcolumn">
      {props.room.players.map((player: Player, idx: number) => {
        return <span key={idx} style={{maxHeight: props.height, color: getPlayerColor(idx).toCSS(true)}}>{player.name}</span>
      })}
    </div>
  )
}

interface CanvasProps {
  room: Room
}

interface CanvasState {
  width: number
  height: number
  loaded: boolean
}

class Canvas extends React.Component<CanvasProps, CanvasState> {
  canvasRef: RefObject<HTMLCanvasElement>
  layer: paper.Layer | undefined

  constructor(props: CanvasProps) {
    super(props)
    this.canvasRef = createRef<HTMLCanvasElement>()
    this.state = {
      width: 0,
      height: 0,
      loaded: false,
    }
  }

  setupBoard = (img: HTMLImageElement) => {
    let raster = new paper.Raster(img)
    raster.position = new paper.Point(raster.size.width / 2, raster.size.height / 2)

    let locations = this.props.room.board.locations
    for (let [idx, loc] of locations.entries()) {
      let c = new Path.Circle(new Point(loc.x, loc.y), 5)
      c.fillColor = new paper.Color("green")

      if (idx > locations.length - 2) {
        continue
      }
      let next_loc = locations[idx+1]
      let l = new Path.Line(new Point(loc.x, loc.y), new Point(next_loc.x, next_loc.y))
      l.strokeColor = new paper.Color("green")
    }

    this.layer = new paper.Layer()
    this.layer.activate()
  }

  drawBoard = () => {
    if (!this.props.room) {
      return
    }
    this.layer?.removeChildren()
    for (let [idx, player] of this.props.room.players.entries()) {
      let color = getPlayerColor(idx)
      let location = this.props.room.board.locations[0]
      for (let loccandidate of this.props.room.board.locations) {
        if (loccandidate.name === player.location) {
          location = loccandidate
          break
        }
      }
      let pcircle = new Path.Circle(new Point(location.x, location.y + idx*5), 10)
      pcircle.fillColor = color
    }
  };

  onImageLoad = () => {
    let img = document.getElementById('baseimg') as HTMLImageElement | null;
    let canvas = document.getElementById('canvas') as HTMLCanvasElement | null;
    if (!img || !canvas) {
      return
    }
    this.setState((prevState) => {
      return {
        width: img?.naturalWidth || 0,
        height: img?.naturalHeight || 0,
        loaded: true
      }
    })
    paper.setup(canvas)
    this.setupBoard(img)
  }

  render = () => {
    if (this.state.loaded) {
      this.drawBoard()
    }
    return (
      <div>
        <img src={serverURL + "/api/board"} onLoad={this.onImageLoad} alt="game board hello" id="baseimg" style={{"display": "none"}} />
        <div className="Flexrow">
          <Playerlist room={this.props.room} height={this.state.height} />
          <canvas ref={this.canvasRef} {...this.props} id="canvas" width={this.state.width} height={this.state.height} />
        </div>
      </div>
    )
  }
}

export default Canvas;