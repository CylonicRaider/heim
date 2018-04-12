import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import classNames from 'classnames'


export default createReactClass({
  displayName: 'CheckField',

  propTypes: {
    name: PropTypes.string.isRequired,
    value: PropTypes.bool,
    onModify: PropTypes.func,
    onValidate: PropTypes.func,
    className: PropTypes.string,
    tabIndex: PropTypes.number,
    disabled: PropTypes.bool,
    children: PropTypes.node,
  },

  onChange(ev) {
    this.props.onModify(ev.target.checked)
  },

  focus() {
    this.refs.input.focus()
  },

  render() {
    return (
      <div className={classNames('check-field', this.props.className)}>
        <input
          ref="input"
          type="checkbox"
          tabIndex={this.props.tabIndex}
          name={this.props.name}
          id={'field-' + this.props.name}
          disabled={this.props.disabled}
          checked={this.props.value || false}
          onChange={this.onChange}
        />
        <label htmlFor={'field-' + this.props.name}>{this.props.children}</label>
      </div>
    )
  },
})
