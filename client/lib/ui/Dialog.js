import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import classNames from 'classnames'

import Popup from './Popup'
import FastButton from './FastButton'
import Spinner from './Spinner'

export default createReactClass({
  displayName: 'Dialog',

  propTypes: {
    title: PropTypes.string,
    working: PropTypes.bool,
    onClose: PropTypes.func,
    className: PropTypes.string,
    children: PropTypes.node,
  },

  onShadeClick(ev) {
    if (ev.target === this.refs.shade) {
      this.props.onClose()
    }
  },

  render() {
    /* eslint-disable jsx-a11y/click-events-have-key-events */
    return (
      <div className="dim-shade dialog-cover fill" ref="shade" onClick={this.onShadeClick}>
        <Popup className={classNames('dialog', this.props.className)}>
          <div className="top-line">
            <div className="logo">
              <div className="emoji emoji-euphoria" />
              euphoria
            </div>
            <div className="title">{this.props.title}</div>
            <Spinner visible={this.props.working} />
            <div className="spacer" />
            <FastButton className="close" onClick={this.props.onClose} />
          </div>
          {this.props.children}
        </Popup>
      </div>
    )
  },
})
