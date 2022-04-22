import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import classNames from 'classnames'
import Entropizer from 'entropizer'

import TextField from './TextField'

const entropizer = new Entropizer()

export default createReactClass({
  displayName: 'PasswordField',

  propTypes: {
    name: PropTypes.string.isRequired,
    label: PropTypes.string.isRequired,
    minEntropy: PropTypes.number,
    showFeedback: PropTypes.bool,
    value: PropTypes.object,
    onModify: PropTypes.func,
    onValidate: PropTypes.func,
    error: PropTypes.bool,
    message: PropTypes.string,
    className: PropTypes.string,
    tabIndex: PropTypes.number,
    disabled: PropTypes.bool,
  },

  getInitialState() {
    return {
      focused: false,
      strength: null,
      message: null,
    }
  },

  componentWillReceiveProps(nextProps) {
    if (this.props.minEntropy !== nextProps.minEntropy) {
      this._checkStrength(this.props.value && this.props.value.text, nextProps.minEntropy)
    }
  },

  onFocus() {
    this.setState({focused: true})
  },

  onBlur() {
    this.setState({focused: false})
  },

  onModify(value) {
    const strength = this._checkStrength(value, this.props.minEntropy)
    this.props.onModify({
      text: value,
      strength: strength,
    })
  },

  _checkStrength(value, minEntropy) {
    let strength = 'unknown'
    if (minEntropy) {
      const entropy = entropizer.evaluate(value)
      let message
      if (entropy < minEntropy) {
        strength = 'weak'
        message = 'too simple — add more!'
      } else {
        strength = 'ok'
      }
      this.setState({strength, message})
    }
    return strength
  },

  focus() {
    this.refs.field.focus()
  },

  render() {
    const strengthClass = this.props.showFeedback ? this.state.strength : null
    const strengthMessage = this.props.showFeedback ? this.state.message : null
    let message
    if (!this.props.message || this.state.focused && strengthMessage) {
      message = strengthMessage
    } else {
      message = this.props.message
    }
    return (
      <TextField
        ref="field"
        inputType="password"
        {...this.props}
        value={this.props.value && this.props.value.text}
        className={classNames('password-field', strengthClass)}
        message={message}
        onModify={this.onModify}
        onFocus={this.onFocus}
        onBlur={this.onBlur}
      />
    )
  },
})
