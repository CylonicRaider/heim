import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import { CSSTransitionGroup } from 'react-transition-group'

module.exports = createReactClass({
  displayName: 'Spinner',

  propTypes: {
    visible: PropTypes.bool,
  },

  getDefaultProps() {
    return {visible: true}
  },

  render() {
    return <CSSTransitionGroup transitionName="spinner" transitionEnterTimeout={100} transitionLeaveTimeout={100}>{this.props.visible && <div key="spinner" className="spinner" />}</CSSTransitionGroup>
  },
})
