import React from 'react'
import createReactClass from 'create-react-class'
import classNames from 'classnames'

import FastButton from './FastButton'


const roomNameRe = /^(pm:)?[a-z0-9]+$/

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
    if (ev.key === 'Escape') {
      this.reset()
    }
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
      // Only reset after the browser acted upon our link
      setImmediate(this.reset)
    } else {
      ev.preventDefault()
    }
  },

  reset() {
    this.setState(this.getInitialState())
  },

  toggle() {
    this.setState({expanded: !this.state.expanded})
  },

  render() {
    const url = this.state.valid ? '/room/' + this.state.text + '/' : '#'
    // Note that valid is a tri-state value, with null being neither valid nor invalid
    return (
      <form className={classNames('room-switcher', this.state.expanded && 'expanded')} action={url} target="_blank" onSubmit={this.apply}>
        <FastButton fastTouch type="button" className={this.state.expanded ? 'room-switcher-cancel' : 'room-switcher-expand'} title="go to another room" onClick={this.toggle} />
        {this.state.expanded && <span className="room-switcher-prompt">go to</span>}
        {this.state.expanded && <span className={classNames('room-switcher-inner', this.state.valid === true && 'valid', this.state.valid === false && 'invalid')}>
          &amp;<input type="text" autoFocus className="room-switcher-room" value={this.state.text} onKeyDown={this.onKeyDown} onChange={this.onChange} />
        </span>}
        {this.state.expanded && <FastButton fastTouch type="submit" className="room-switcher-apply" />}
      </form>
    )
  },
})
