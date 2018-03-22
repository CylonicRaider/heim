import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import ReactDOM from 'react-dom'
import { CSSTransitionGroup } from 'react-transition-group'
import classNames from 'classnames'

import Popup from './Popup'


export default createReactClass({
  displayName: 'Bubble',

  propTypes: {
    visible: PropTypes.bool,
    anchorEl: PropTypes.any,
    className: PropTypes.string,
    transition: PropTypes.string,
    offset: PropTypes.func,
    edgeSpacing: PropTypes.number,
    onDismiss: PropTypes.func,
    children: PropTypes.node,
  },

  mixins: [require('react-immutable-render-mixin')],

  getDefaultProps() {
    return {
      edgeSpacing: 10,
      transition: 'slide-down',
    }
  },

  componentDidMount() {
    this.reposition()
  },

  componentDidUpdate() {
    this.reposition()
  },

  onDismiss(ev) {
    if (this.props.visible && this.props.onDismiss) {
      this.props.onDismiss(ev)
    }
  },

  reposition() {
    // FIXME: only handles left anchors. expand/complexify to work for multiple
    // orientations when necessary.
    if (this.props.visible && this.props.anchorEl) {
      const box = this.props.anchorEl.getBoundingClientRect()
      const node = ReactDOM.findDOMNode(this.refs.bubble)

      let top = box.top
      top -= Math.max(0, top + node.clientHeight + this.props.edgeSpacing - uiwindow.innerHeight)

      let left = box.right

      if (this.props.offset) {
        const offsetBox = this.props.offset(box)
        left -= offsetBox.left || 0
        top -= offsetBox.top || 0
      }

      node.style.left = left + 'px'
      node.style.top = top + 'px'
    }
  },

  render() {
    return (
      <CSSTransitionGroup transitionName={this.props.transition} transitionEnterTimeout={150} transitionLeaveTimeout={150}>
        {this.props.visible &&
          <Popup ref="bubble" key="bubble" className={classNames('bubble', this.props.className)} onDismiss={this.onDismiss}>
            {this.props.children}
          </Popup>
        }
      </CSSTransitionGroup>
    )
  },
})
