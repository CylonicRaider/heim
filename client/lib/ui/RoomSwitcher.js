import React from 'react'
import createReactClass from 'create-react-class'
import classNames from 'classnames'

import FastButton from './FastButton'


const roomNameRe = /(pm:)?[a-z0-9]+/

export default createReactClass({
  displayName: 'RoomSwitcher',

  getInitialState() {
    return {
      expanded: false,
      text: '',
      valid: null,
    }
  },

  onKeyDown(ev) {
    ev.stopPropagation()
  },

  onChange(ev) {
    const newText = ev.target.value
    const newValid = newText ? roomNameRe.test(newText) : null
    this.setState({text: newText, valid: newValid})
    ev.stopPropagation()
  },

  apply(ev) {
    if (this.state.valid) {
      this.setState(this.getInitialState())
    } else {
      ev.preventDefault()
    }
  },

  toggle() {
    this.setState({expanded: !this.state.expanded})
  },

  render() {
    // Note that valid is a tri-state value, with null being neither valid nor invalid
    return (
      <span className={classNames('room-switcher', this.state.expanded && 'expanded')}>
        <FastButton fastTouch className={this.state.expanded ? 'room-switcher-cancel' : 'room-switcher-expand'} title="go to another room" onClick={this.toggle} />
        {this.state.expanded && <span className="room-switcher-prompt">go to</span>}
        {this.state.expanded && <span className={classNames('room-switcher-inner', this.state.valid === true && 'valid', this.state.valid === false && 'invalid')}>
          &amp;<input autoFocus className="room-switcher-room" value={this.state.text} onKeyDown={this.onKeyDown} onChange={this.onChange} />
        </span>}
        {this.state.expanded && <FastButton ref="link" component="a" href={this.state.valid ? '/room/' + this.state.text + '/' : '#'} target="_blank" fastTouch className="room-switcher-apply" onClick={this.apply} />}
      </span>
    )
  },
})
