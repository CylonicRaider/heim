import _ from 'lodash'
import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import ReactDOM from 'react-dom'
import Reflux from 'reflux'
import Immutable from 'immutable'

import chat from '../stores/chat'
import storage from '../stores/storage'
import FastButton from '../ui/FastButton'
import Bubble from '../ui/Bubble'
import CheckField from '../ui/forms/CheckField'


const storeActions = Reflux.createActions([
  'setTheme',
  'showAllReplies',
  'showThemeDialog',
])
_.extend(module.exports, storeActions)

export const store = Reflux.createStore({
  listenables: [
    storeActions,
    {chatChange: chat.store},
    {storageChange: storage.store},
  ],

  init() {
    this.state = Immutable.fromJS({theme: null, showAllReplies: false, dialogVisible: false})
    this.chatState = null
    this.storageState = null
  },

  getInitialState() {
    return this.state
  },

  setTheme(newTheme) {
    this._updateState(this.state.set('theme', newTheme))
  },

  showAllReplies(newState) {
    this._updateState(this.state.set('showAllReplies', newState))
  },

  showThemeDialog(newState) {
    this._updateState(this.state.set('dialogVisible', newState))
  },

  _updateState(newState) {
    if (newState.equals(this.state)) {
      return
    }

    if (newState.get('theme') !== this.state.get('theme')) {
      storage.setRoom(this.chatState.roomName, 'theme', newState.get('theme'))
    }
    if (newState.get('showAllReplies') !== this.state.get('showAllReplies')) {
      storage.setRoom(this.chatState.roomName, 'showAllReplies', newState.get('showAllReplies'))
    }

    this.state = newState
    this.trigger(this.state)
  },

  _sync() {
    // Poll of the user base resulted in a total of one valid vote, which was in favor of room-per-room settings.
    if (this.storageState && this.chatState && this.chatState.roomName) {
      const roomData = this.storageState.room[this.chatState.roomName]
      if (roomData) {
        this._updateState(this.state.merge({theme: roomData.theme, showAllReplies: roomData.showAllReplies}))
      }
    }
  },

  chatChange(state) {
    this.chatState = state
    this._sync()
  },

  storageChange(data) {
    this.storageState = data
    this._sync()
  },
})

export const ThemeChooserButton = createReactClass({
  displayName: 'ThemeChooserButton',

  mixins: [
    Reflux.connect(store, 'settings')
  ],

  toggleSettings() {
    storeActions.showThemeDialog(!this.state.settings.get('dialogVisible'))
  },

  render() {
    return (
      <FastButton fastTouch className="theme-button" onClick={this.toggleSettings}>
        theme
      </FastButton>
    )
  },
})

export const ThemeChooserDialog = createReactClass({
  displayName: 'ThemeChooserDialog',

  propTypes: {
    anchorEl: PropTypes.any,
  },

  mixins: [
    Reflux.connect(store, 'settings')
  ],

  dismiss(ev) {
    // ensure the bubble can be closed by clicking the button
    setImmediate(() => {
      storeActions.showThemeDialog(false)
    })
    if (ev) {
      ev.stopPropagation()
    }
  },

  _updateAnchor() {
    this.anchorEl = this.props.anchor && ReactDOM.findDOMNode(this.props.anchor)
  },

  componentDidMount() {
    this._updateAnchor()
  },

  componentDidUpdate() {
    this._updateAnchor()
  },

  onShowAllReplies(ev) {
    storeActions.showAllReplies(ev.target.checked)
  },

  render() {
    return (
      <Bubble transition="slide-right" visible={this.state.settings.get('dialogVisible')} anchorEl={this.anchorEl} onDismiss={this.dismiss}>
        <div>
          <input type="checkbox" checked={this.state.settings.get('showAllReplies')} onChange={this.onShowAllReplies} id="theme-showAllReplies"/>
          <label htmlFor="theme-showAllReplies">Show all replies</label>
        </div>
      </Bubble>
    )
  },
})

export function install() {
  let buttonComp

  Heim.hook('info-pane', () => (
    <ThemeChooserButton key="theme-chooser-button" ref={(el) => { buttonComp = el }} />
  ))

  Heim.hook('page-bottom', () => (
    <ThemeChooserDialog key="theme-chooser-dialog" anchor={buttonComp} />
  ))
}
