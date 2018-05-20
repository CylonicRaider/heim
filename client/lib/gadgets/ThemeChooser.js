import _ from 'lodash'
import Reflux from 'reflux'
import Immutable from 'immutable'

import chat from '../stores/chat'
import storage from '../stores/storage'


const storeActions = Reflux.createActions([
  'setTheme',
])
_.extend(module.exports, storeActions)

export const store = Reflux.createStore({
  listenables: [
    storeActions,
    {chatChange: chat.store},
    {storageChange: storage.store},
  ],

  init() {
    this.state = Immutable.fromJS({theme: null})
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

  _update() {
    // Poll of the user base resulted in a total of one valid vote, which was in favor of room-per-room settings.
    if (this.storageState && this.chatState && this.chatState.roomName) {
      this.setTheme(this.storageState.room[this.chatState.roomName].theme)
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
