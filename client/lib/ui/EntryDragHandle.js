import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import classNames from 'classnames'
import Reflux from 'reflux'

import ui from '../stores/ui'

export default createReactClass({
  displayName: 'EntryDragHandle',

  propTypes: {
    pane: PropTypes.instanceOf(ui.Pane).isRequired,
  },

  mixins: [
    Reflux.ListenerMixin,
  ],

  getInitialState() {
    return {
      pane: this.props.pane.store.getInitialState(),
    }
  },

  componentDidMount() {
    this.listenTo(this.props.pane.store, (state) => this.setState({'pane': state}))
  },

  render() {
    const showJumpToBottom = this.state.pane.draggingEntry && this.state.pane.focusedMessage !== null
    return (
      <div className="drag-handle-container">
        <button type="button" className={classNames('drag-handle', {'touching': this.state.pane.draggingEntry})} />
        {showJumpToBottom && <button type="button" className={classNames('jump-to-bottom', {'touching': this.state.pane.draggingEntryCommand === 'to-bottom'})} />}
      </div>
    )
  },
})
