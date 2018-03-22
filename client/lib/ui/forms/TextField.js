import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import classNames from 'classnames'

import FieldLabelContainer from './FieldLabelContainer'


export default createReactClass({
  displayName: 'TextField',

  propTypes: {
    name: PropTypes.string.isRequired,
    label: PropTypes.string.isRequired,
    value: PropTypes.string,
    onModify: PropTypes.func,
    onValidate: PropTypes.func,
    onFocus: PropTypes.func,
    onBlur: PropTypes.func,
    action: PropTypes.string,
    onAction: PropTypes.func,
    error: PropTypes.bool,
    isFirstError: PropTypes.bool,
    autoFocus: PropTypes.bool,
    message: PropTypes.string,
    className: PropTypes.string,
    inputType: PropTypes.string,
    tabIndex: PropTypes.number,
    spellCheck: PropTypes.bool,
    disabled: PropTypes.bool,
  },

  componentDidUpdate(prevProps) {
    if (!prevProps.error && this.props.isFirstError) {
      this.refs.input.focus()
    }
  },

  onChange(ev) {
    this.props.onModify(ev.target.value)
  },

  onFocus(ev) {
    if (this.props.onFocus) {
      this.props.onFocus(ev)
    }
  },

  onBlur(ev) {
    if (!this.props.disabled) {
      this.props.onValidate()
    }
    if (this.props.onBlur) {
      this.props.onBlur(ev)
    }
  },

  focus() {
    this.refs.input.focus()
  },

  render() {
    return (
      <FieldLabelContainer
        className={classNames('text-field', this.props.className)}
        label={this.props.label}
        action={this.props.action}
        onAction={this.props.onAction}
        error={this.props.error}
        message={this.props.message}
      >
        <input
          ref="input"
          name={this.props.name}
          type={this.props.inputType}
          value={this.props.value}
          tabIndex={this.props.tabIndex}
          autoFocus={this.props.autoFocus}
          spellCheck={this.props.spellCheck}
          disabled={this.props.disabled}
          onChange={this.onChange}
          onFocus={this.onFocus}
          onBlur={this.onBlur}
        />
      </FieldLabelContainer>
    )
  },
})
