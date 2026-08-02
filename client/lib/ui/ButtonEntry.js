import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import ReactDOM from 'react-dom'
import classNames from 'classnames'
import Reflux from 'reflux'

import ui from '../stores/ui'
import EntryDragHandle from './EntryDragHandle'
import FastButton from './FastButton'
import KeyboardActionHandler from './KeyboardActionHandler'

export default createReactClass({
  displayName: 'ButtonEntry',

  propTypes: {
    className: PropTypes.oneOfType([PropTypes.string, PropTypes.array, PropTypes.object]),
    pane: PropTypes.instanceOf(ui.Pane).isRequired,
    onClick: PropTypes.func,
    keys: PropTypes.objectOf(PropTypes.func),
    focused: PropTypes.bool,
  },

  mixins: [
    Reflux.ListenerMixin,
  ],

  componentDidMount() {
    this.listenTo(this.props.pane.focusEntry, 'onPaneFocus')
    this.listenTo(this.props.pane.blurEntry, 'onPaneBlur')
  },

  onClick(ev) {
    if (this.props.onClick) {
      this.props.onClick(ev)
    }
  },

  onPaneFocus() {
    if (this.props.focused) {
      ReactDOM.findDOMNode(this.refs.button).focus()
    }
  },

  onPaneBlur() {
    if (this.props.focused) {
      ReactDOM.findDOMNode(this.refs.button).blur()
    }
  },

  render() {
    const pane = this.props.pane
    const focused = this.props.focused

    /* eslint-disable jsx-a11y/click-events-have-key-events */
    let ret = <FastButton ref="button" component="div" className={classNames(this.props.className, {'focus-target': focused})} tabIndex={0} onClick={this.onClick}>
      {this.props.children}
      {focused && <div className="spacer"><EntryDragHandle pane={this.props.pane} /></div>}
    </FastButton>
    if (focused) {
      ret = <KeyboardActionHandler listenTo={pane.keydownOnPane} keys={{
        ArrowLeft: () => pane.moveMessageFocus('out'),
        ArrowRight: () => pane.moveMessageFocus('top'),
        ArrowUp: () => pane.moveMessageFocus('up'),
        ArrowDown: () => pane.moveMessageFocus('down'),
        Enter: this.onClick,
        Escape: () => pane.escape(),
        // avoid keyboard focus escaping and wreaking havoc
        Tab: () => true,
        ShiftTab: () => true,
       ...this.props.keys,
      }}>
        {ret}
      </KeyboardActionHandler>
    }
    return ret
  },
})
