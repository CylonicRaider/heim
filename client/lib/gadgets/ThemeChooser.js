/* eslint-disable react/no-multi-comp */

import _ from 'lodash'
import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import classNames from 'classnames'
import ReactDOM from 'react-dom'
import Reflux from 'reflux'
import Immutable from 'immutable'

import chat from '../stores/chat'
import ui from '../stores/ui'
import storage from '../stores/storage'
import FastButton from '../ui/FastButton'
import Bubble from '../ui/Bubble'
import heimURL from '../heimURL'
import { CheckBox, RadioBox } from './formControls'


const themes = ['default', 'dark', 'spooky']
const themeNames = {default: 'Default', dark: 'Dark', spooky: 'Spooky'}

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
    this.state = Immutable.fromJS({theme: 'default', showAllReplies: false, dialogVisible: false})
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
        const newTheme = roomData.theme || 'default'
        this._updateState(this.state.merge({theme: newTheme, showAllReplies: roomData.showAllReplies}))
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
    Reflux.connect(store, 'settings'),
    Reflux.connect(ui.store, 'ui'),
  ],

  toggleSettings() {
    storeActions.showThemeDialog(!this.state.settings.get('dialogVisible'))
  },

  render() {
    return (
      <FastButton fastTouch className={classNames('theme-chooser-button', this.state.ui.thin && 'thin-ui')} onClick={this.toggleSettings}>
        <span ref="anchor" className="anchor" />
        theme
      </FastButton>
    )
  },
})

export const ThemeChooserDialog = createReactClass({
  displayName: 'ThemeChooserDialog',

  propTypes: {
    anchor: PropTypes.any,
  },

  mixins: [
    Reflux.connect(store, 'settings'),
    Reflux.connect(ui.store, 'ui'),
  ],

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

  _updateAnchor() {
    this.anchorEl = this.props.anchor && ReactDOM.findDOMNode(this.props.anchor)
  },

  dismiss(ev) {
    // ensure the bubble can be closed by clicking the button
    setImmediate(() => {
      storeActions.showThemeDialog(false)
    })
    if (ev) {
      ev.stopPropagation()
    }
  },

  render() {
    return (
      <Bubble className="theme-chooser-dialog" transition={this.state.ui.thin ? 'slide-down' : 'slide-right'} visible={this.state.settings.get('dialogVisible')} anchorEl={this.anchorEl} onDismiss={this.dismiss}>
        <div className="field-group">
          <span className="field-group-label">Theme:</span>
          {themes.map(name => (
            <RadioBox name="theme" key={name} value={name} checked={this.state.settings.get('theme') === name} onChange={this.onThemeChange}>{themeNames[name]}</RadioBox>
          ))}
        </div>
        <hr className="separator" />
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
    <ThemeChooserDialog key="theme-chooser-dialog" anchor={buttonComp && buttonComp.refs.anchor} />
  ))

  Heim.hook('page-bottom', () => {
    const theme = store.state.get('theme')
    if (theme == null || theme === 'default') return null // eslint-disable-line eqeqeq
    return (
      <link key="user-theme" rel="stylesheet" type="text/css" href={heimURL('/static/theme-' + theme + '.css')} />
    )
  })
}
