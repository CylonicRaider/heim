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
    } else if (ev.key === 'Enter') {
      // HACK: React defers hadling this event, such that it happens outside of a user interaction context,
      //       such that trying to open in a new tab triggers popup blocking. Hence, we do *not* intercept
      //       Enter-with-no-Shift here, ensure the new-tab button is the form's default submit button,
      //       and wait for the form submit event instead.
      if (ev.shiftKey) {
        this.apply(false)
        ev.preventDefault()
      }
    }
    ev.stopPropagation()
  },

  onChange(ev) {
    const newText = ev.target.value
    const newValid = newText ? roomNameRe.test(newText) : null
    this.setState({text: newText, valid: newValid})
    ev.stopPropagation()
  },

  onSubmit(ev) {
    this.apply(!(ev.nativeEvent.submitter && ev.nativeEvent.submitter.name === 'apply-here'))
    ev.preventDefault()
  },

  apply(newTab) {
    if (!this.state.valid) return false

    const url = '/room/' + this.state.text + '/' + uiwindow.location.search + uiwindow.location.hash
    if (newTab) {
      uiwindow.open(url, '_blank')
    } else {
      uiwindow.location.href = url
    }

    this.reset()
    return true
  },

  reset() {
    this.setState(this.getInitialState())
  },

  toggle() {
    /* eslint-disable react/no-access-state-in-setstate */
    this.setState({expanded: !this.state.expanded})
  },

  render() {
    const url = this.state.valid ? '/room/' + this.state.text + '/' : '#'
    // Note that valid is a tri-state value, with null being neither valid nor invalid
    return (
      <form className={classNames('room-switcher', this.state.expanded && 'expanded')} action={url} target="_blank" onSubmit={this.onSubmit}>
        <FastButton
          fastTouch
          type="button"
          className={this.state.expanded ? 'room-switcher-cancel' : 'room-switcher-expand'}
          title={this.state.expanded ? 'do not go to another room' : 'go to another room'}
          onClick={this.toggle}
        />
        {this.state.expanded && <span className="room-switcher-prompt">go to</span>}
        {this.state.expanded && (
          <span className={classNames('room-switcher-inner', this.state.valid === true && 'valid', this.state.valid === false && 'invalid')}>
            &amp;<input type="text" autoFocus className="room-switcher-room" value={this.state.text} onKeyDown={this.onKeyDown} onChange={this.onChange} />
          </span>
        )}
        {/* WARNING: Implicitly submitting the form (by pressing Enter in the input) must trigger the new-tab button; see onKeyDown(). */}
        {this.state.expanded && <FastButton fastTouch type="submit" title="open room in new tab" name="apply" className="room-switcher-apply" />}
        {this.state.expanded && <FastButton fastTouch type="submit" formtarget="_top" title="open room in this tab" name="apply-here" className="room-switcher-apply-here" />}
      </form>
    )
  },
})
