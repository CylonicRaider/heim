import _ from 'lodash'
import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import Reflux from 'reflux'

import forwardProps from '../forwardProps'


export default createReactClass({
  displayName: 'KeyboardActionHandler',

  propTypes: {
    listenTo: PropTypes.func,
    keys: PropTypes.objectOf(PropTypes.func),
    children: PropTypes.node,
  },

  mixins: [
    Reflux.ListenerMixin,
  ],

  componentDidMount() {
    this.listenTo(this.props.listenTo, 'onKeyDown')
  },

  onKeyDown(ev) {
    let key = ev.key

    if (ev.ctrlKey) {
      key = 'Control' + key
    }

    if (ev.altKey) {
      key = 'Alt' + key
    }

    if (ev.shiftKey) {
      key = 'Shift' + key
    }

    if (ev.metaKey) {
      key = 'Meta' + key
    }

    if (key !== 'Tab' && Heim.tabPressed) {
      key = 'Tab' + key
    }

    const handler = this.props.keys[key]
    if (handler && handler(ev) !== false) {
      ev.stopPropagation()
      ev.preventDefault()
    }
  },

  render() {
    return (
      <div onKeyDown={this.onKeyDown} {...forwardProps(this)}>
        {this.props.children}
      </div>
    )
  },
})
