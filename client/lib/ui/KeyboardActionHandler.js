import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import Reflux from 'reflux'

import forwardProps from '../util/forwardProps'

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

    if (ev.ctrlKey && ev.key !== 'Control') {
      key = 'Control' + key
    }

    if (ev.altKey && ev.key !== 'Alt') {
      key = 'Alt' + key
    }

    if (ev.shiftKey && ev.key !== 'Shift') {
      key = 'Shift' + key
    }

    if (ev.metaKey && ev.key !== 'Meta') {
      key = 'Meta' + key
    }

    if (Heim.tabPressed && ev.key !== 'Tab') {
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
      <div onKeyDown={this.onKeyDown} {...forwardProps(this, /^on/)}>
        {this.props.children}
      </div>
    )
  },
})
