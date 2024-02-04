import _ from 'lodash'
import Reflux from 'reflux'

import Hooks from '../Hooks'
import fauxPlugins from '../fauxPlugins'

const storeActions = Reflux.createActions([
  'load',
  'apply',
  'applySingle',
])
_.extend(module.exports, storeActions)

const hooks = module.exports.hooks = new Hooks(
  'page-bottom',
  'main-sidebar',
  'thread-panes',
  'incoming-messages',
  'main-pane-top',
  'top-bar-middle',
  'info-pane'
)

module.exports.hook = hooks.register.bind(hooks)

module.exports.store = Reflux.createStore({
  listenables: storeActions,

  init() {
    this.state = {applied: {}}
  },

  getInitialState() {
    return this.state
  },

  load(roomName) {
    fauxPlugins(roomName)
  },

  _applySingle(Func) {
    try {
      this.state.applied[Func.name] = new Func(Heim)
    } catch (exc) {
      console.error('Could not apply plugin:', exc)  // eslint-disable-line no-console
    }
  },

  apply(list) {
    list.forEach((Func) => this._applySingle(Func))
    this.trigger(this.state)
  },

  applySingle(Func) {
    this._applySingle(Func)
    this.trigger(this.state)
  },
})

module.exports.linkLocal = (locals) => {
  if (locals._onAdd === storeActions.applySingle) {
    return
  }
  storeActions.apply(locals.pluginList)
  locals._onAdd = storeActions.applySingle
}

module.exports.unlinkLocal = (locals) => {
  if (locals._onAdd !== storeActions.applySingle) {
    return
  }
  locals._onAdd = null
}
