import _ from 'lodash'
import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import Reflux from 'reflux'

import actions from '../actions'
import chat from '../stores/chat'
import ui from '../stores/ui'
import EntryMixin from './EntryMixin'

export default createReactClass({
  displayName: 'PasscodeEntry',

  propTypes: {
    pane: PropTypes.instanceOf(ui.Pane).isRequired,
  },

  mixins: [
    EntryMixin,
    Reflux.listenTo(chat.store, '_onChatUpdate'),
  ],

  getInitialState() {
    return {
      value: '',
      connected: null,
      authState: null,
    }
  },

  componentWillMount() {
    // debounce state changes to reduce jank from fast responses
    // TODO: break out into a debounced connect mixin, once chat store is fully immutable?
    this._onChatUpdate = _.debounce(this.onChatUpdate, 250, {leading: true, trailing: true})
  },

  componentDidMount() {
    this.listenTo(this.props.pane.focusEntry, 'focus')
    this.listenTo(this.props.pane.blurEntry, 'blur')
    this.listenTo(this.props.pane.keydownOnPane, 'proxyKeyDown')
  },

  onChatUpdate(chatState) {
    this.setState({
      connected: chatState.connected,
      authState: chatState.authState,
    })
  },

  tryPasscode(ev) {
    this.refs.input.focus()
    ev.preventDefault()

    if (this.state.authState === 'trying') {
      return
    }

    actions.tryRoomPasscode(this.state.value)
    this.setState({value: ''})
  },

  render() {
    let label
    switch (this.state.authState) {
      case 'trying':
        label = 'trying...'
        break
      case 'failed':
        label = 'no dice. try again:'
        break
      default:
        label = 'passcode:'
    }

    return (
      /* eslint-disable jsx-a11y/label-has-associated-control */
      <div className="entry-box passcode">
        <p className="message">This room requires a passcode.</p>
        <form className="entry focus-target" onSubmit={this.tryPasscode}>
          <label htmlFor="passcode-entry">{label}</label>
          <input key="passcode" ref="input" type="password" id="passcode-entry" className="entry-text" autoFocus value={this.state.value} onChange={(event) => this.setState({value: event.target.value})} disabled={this.state.connected === false} />
        </form>
      </div>
    )
  },
})
