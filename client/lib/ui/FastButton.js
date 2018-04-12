import _ from 'lodash'
import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import ReactDOM from 'react-dom'

import forwardProps from '../forwardProps'

// A button that triggers on touch start on mobile to increase responsiveness.
export default createReactClass({
  displayName: 'FastButton',

  propTypes: {
    vibrate: PropTypes.bool,
    disabled: PropTypes.bool,
    fastTouch: PropTypes.bool,
    empty: PropTypes.bool,
    onClick: PropTypes.func,
    component: PropTypes.string,
    children: PropTypes.node,
    // for default prop
    tabIndex: PropTypes.number,
  },

  getDefaultProps() {
    return {
      component: 'button',
      tabIndex: 0,
    }
  },

  onClick(ev) {
    if (Heim.isTouch) {
      if (ev.type === 'touchstart') {
        if (this.props.vibrate && !this.props.disabled && Heim.isAndroid && navigator.vibrate) {
          navigator.vibrate(7)
        }

        if (this.props.fastTouch) {
          // prevent emulated click event
          ev.preventDefault()
        } else {
          return
        }
      }
    }

    if (this.props.onClick) {
      this.props.onClick(ev)
    }
  },

  onTouchStart(ev) {
    ReactDOM.findDOMNode(this).classList.add('touching')
    this.onClick(ev)
  },

  onTouchEnd() {
    ReactDOM.findDOMNode(this).classList.remove('touching')
  },

  onKeyDown(ev) {
    if (ev.key === 'Enter' || ev.key === 'Space') {
      this.props.onClick(ev)
    }
  },

  render() {
    // https://bugzilla.mozilla.org/show_bug.cgi?id=984869#c2
    return React.createElement(
      this.props.component,
      _.extend(forwardProps(this), {
        onClick: this.onClick,
        onTouchStart: this.onTouchStart,
        onTouchEnd: this.onTouchEnd,
        onTouchCancel: this.onTouchEnd,
        onKeyDown: this.onKeyDown,
      }),
      !this.props.empty && <div className="inner">{this.props.children}</div>
    )
  },
})
