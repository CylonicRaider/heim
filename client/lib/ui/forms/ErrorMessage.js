import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import classNames from 'classnames'


export default createReactClass({
  displayName: 'ErrorMessage',

  propTypes: {
    name: PropTypes.string.isRequired,
    message: PropTypes.string,
    error: PropTypes.bool,
    className: PropTypes.string,
  },

  render() {
    return <div className={classNames('message', this.props.error && 'error', this.props.className)}>{this.props.message}</div>
  },
})
