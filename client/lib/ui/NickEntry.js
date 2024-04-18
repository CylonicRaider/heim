import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import Reflux from 'reflux'

import actions from '../heim/actions'
import ui from '../stores/ui'
import EntryMixin from './EntryMixin'

export default createReactClass({
  displayName: 'NickEntry',

  propTypes: {
    pane: PropTypes.instanceOf(ui.Pane).isRequired,
  },

  mixins: [
    EntryMixin,
    Reflux.ListenerMixin,
    Reflux.connect(require('../stores/chat').store, 'chat'),
  ],

  getInitialState() {
    return {value: ''}
  },

  componentDidMount() {
    this.listenTo(this.props.pane.focusEntry, 'focus')
    this.listenTo(this.props.pane.blurEntry, 'blur')
    this.listenTo(this.props.pane.keydownOnPane, 'proxyKeyDown')
  },

  setNick(ev) {
    this.refs.input.focus()
    ev.preventDefault()

    actions.setNick(this.state.value)
  },

  render() {
    /* eslint-disable jsx-a11y/label-has-associated-control */
    return (
      <div className="entry-box welcome">
        <div className="message">
          <h1><strong>Hello{this.state.value ? ' ' + this.state.value : ''}!</strong> <span className="no-break">Welcome to our discussion.</span></h1>
          <p>To reply to a message directly, {Heim.isTouch ? 'tap' : 'use the arrow keys or click on'} it.</p>
        </div>
        <form className="entry focus-target" onSubmit={this.setNick}>
          <label htmlFor="nick-entry">choose your name to begin:</label>
          <input key="nick" id="nick-entry" ref="input" type="text" className="entry-text" autoFocus value={this.state.value} onChange={(event) => this.setState({value: event.target.value})} disabled={this.state.chat.connected === false} />
        </form>
      </div>
    )
  },
})
