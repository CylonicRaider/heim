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
    this.state = Immutable.fromJS({theme: null, dialogVisible: false})
    this.chatState = null
    this.storageState = null
  },

  getInitialState() {
    return this.state
  },

  setTheme(newTheme) {
    if (newTheme === this.state.get('theme')) {
      return
    }

    storage.store.setRoom(this.chatState.roomName, 'theme', newTheme)

    this.state = this.state.set('theme', newTheme)
    this.trigger(this.state)
  },

  showThemeDialog(state) {
    this.state = this.state.set('dialogVisible', state)
    this.trigger(this.state)
  },

  _update() {
    // Poll of the user base resulted in a total of one valid vote, which was in favor of room-per-room settings.
    if (this.storageState && this.chatState && this.chatState.roomName) {
      const roomData = this.storageState.room[this.chatState.roomName]
      if (roomData) {
        this.setTheme(roomData.theme)
      }
    }
  },

  chatChange(state) {
    this.chatState = state
    this._update()
  },

  storageChange(data) {
    this.storageState = data
    this._update()
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

  dismiss() {
    // ensure the bubble can be closed by clicking the button
    setImmediate(() => {
      storeActions.showThemeDialog(false)
    })
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

  render() {
    return (
      <Bubble transition="slide-right" visible={this.state.settings.get('dialogVisible')} anchorEl={this.anchorEl} onDismiss={this.dismiss}>
        Hello World!
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
