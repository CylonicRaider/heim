import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import classNames from 'classnames'


export const CheckBox = createReactClass({
  displayName: 'CheckBox',

  propTypes: {
    name: PropTypes.string,
    checked: PropTypes.bool,
    onChange: PropTypes.func,
  },

  render() {
    return (
      <label className={classNames('form-control', {'checked': this.props.checked})}>
        <span className="form-control-button form-control-checkbox"><span><input type="checkbox" name={this.props.name} checked={this.props.checked} onChange={this.props.onChange} /></span></span>
        <span className="form-control-text">{this.props.children}</span>
      </label>
    )
  },

})

export const RadioBox = createReactClass({
  displayName: 'RadioBox',

  propTypes: {
    name: PropTypes.string,
    value: PropTypes.string,
    checked: PropTypes.bool,
    onChange: PropTypes.func,
  },

  render() {
    return (
      <label className={classNames('form-control', {'checked': this.props.checked})}>
        <span className="form-control-button form-control-radiobox"><span><input type="radio" name={this.props.name} value={this.props.value} checked={this.props.checked} onChange={this.props.onChange} /></span></span>
        <span className="form-control-text">{this.props.children}</span>
      </label>
    )
  },

})
