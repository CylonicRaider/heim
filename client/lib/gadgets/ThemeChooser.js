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
import heimURL from '../heimURL'
import { CheckBox, RadioBox } from './formControls'


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
      <FastButton fastTouch className="theme-chooser-button" onClick={this.toggleSettings}>
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

  onThemeChange(ev) {
    const theme = ev.target.value ? ev.target.value : null
    storeActions.setTheme(theme)
  },

  render() {
    return (
      <Bubble className="theme-chooser-dialog" transition="slide-right" offset={() => ({ left: 5, top: 0 })} visible={this.state.settings.get('dialogVisible')} anchorEl={this.anchorEl} onDismiss={this.dismiss}>
        <div className="field-group">
          <span className="field-group-label">Theme:</span>
          <RadioBox name="theme" value="" checked={this.state.settings.get('theme') == null} onChange={this.onThemeChange}>Default</RadioBox>
          <RadioBox name="theme" value="base" checked={this.state.settings.get('theme') == 'base'} onChange={this.onThemeChange}>Development</RadioBox>
          <RadioBox name="theme" value="darcula" checked={this.state.settings.get('theme') == 'darcula'} onChange={this.onThemeChange}>Darcula</RadioBox>
          <RadioBox name="theme" value="spooky" checked={this.state.settings.get('theme') == 'spooky'} onChange={this.onThemeChange}>Spooky</RadioBox>
        </div>
        <hr className="spacer" />
        <CheckBox checked={this.state.settings.get('showAllReplies')} onChange={this.onShowAllReplies}>Show all replies</CheckBox>
      </Bubble>
    )
  },
})

export function install(params) {
  let buttonComp

  if (params.theme) {
    // delay setting the theme because doing is synchonously crashes
    setImmediate(() => storeActions.setTheme(params.theme))
  }

  Heim.hook('info-pane', () => (
    <ThemeChooserButton key="theme-chooser-button" ref={(el) => { buttonComp = el }} />
  ))

  Heim.hook('page-bottom', () => (
    <ThemeChooserDialog key="theme-chooser-dialog" anchor={buttonComp} />
  ))

  Heim.hook('page-bottom', () => {
    const theme = store.state.get('theme')
    if (theme == null) return null
    return (
      <link key="user-theme" rel="stylesheet" type="text/css" href={heimURL('/static/theme-' + theme + '.css')} />
    )
  })
}
