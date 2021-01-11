import React from 'react';
import { Room } from './Elements'

interface HistoryProps {
  room?: Room
}

interface HistoryState {
}

class History extends React.Component<HistoryProps,HistoryState> {
  render() {
    return (
      <div className="scrollable card">
        {this.props.room?.history.map((_, index, array) => (
          <div key={array.length - 1 - index}>{array[array.length - 1 - index]}</div>
        ))}
      </div>
    )
  }
}

export default History