/* eslint-disable react/no-access-state-in-setstate */
import _ from 'lodash'
import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import classNames from 'classnames'

export default createReactClass({
  displayName: 'Form',

  propTypes: {
    context: PropTypes.object,
    errors: PropTypes.objectOf(PropTypes.string),
    validators: PropTypes.objectOf(PropTypes.func),
    working: PropTypes.bool,
    onSubmit: PropTypes.func,
    className: PropTypes.string,
    children: PropTypes.node,
  },

  getDefaultProps() {
    return {
      errors: {},
      context: {},
      validators: {},
    }
  },

  getInitialState() {
    return {
      values: {},
      errors: {},
    }
  },

  componentWillMount() {
    this._strict = false
  },

  componentWillReceiveProps(nextProps) {
    if (!_.isEqual(this.props.context, nextProps.context) || !_.isEqual(this.props.validators, nextProps.validators)) {
      this._strict = false
      this.setState({errors: this._validateFields(nextProps.validators, this.state.values, nextProps.context)})
    }
  },

  onFieldModify(name, value) {
    const values = _.assign({}, this.state.values)
    values[name] = value
    this.setState({
      values: values,
      errors: _.assignWith(this.state.errors, this._validateField(name, values), this._clearError),
    })
  },

  onFieldValidate(name) {
    this.setState({
      errors: _.assign(this.state.errors, this._validateField(name, this.state.values)),
    })
  },

  onSubmit(ev) {
    ev.preventDefault()
    this._strict = true
    const errors = this._validateFields(this.props.validators, this.state.values, this.props.context)
    if (!_.some(errors)) {
      this.setState({errors: {}})
      this._strict = false
      this.props.onSubmit(this.state.values)
    } else {
      this.setState({errors: errors})
    }
  },

  _validateFields(validators, formValues, context, filter) {
    const errors = {}
    _.each(validators, (validator, fieldSpec) => {
      if (!validator) {
        return
      }

      const validatorValues = {}
      fieldSpec.split(' ').forEach((field) => {
        validatorValues[field] = formValues[field]
      })
      if (!filter || filter(validatorValues)) {
        _.assign(errors, validator(validatorValues, this._strict, context))
      }
    })
    return errors
  },

  _validateField(name, formValues) {
    return this._validateFields(this.props.validators, formValues, this.props.context, (values) => values.hasOwnProperty(name))
  },

  _clearError(origError, newError) {
    return !newError ? null : origError
  },

  _walkChildren(children, serverErrors, validatorErrors, errorSeen) {
    let foundError = errorSeen
    const errors = _.assign({}, serverErrors, validatorErrors)
    return React.Children.map(children, (child) => {
      if (!React.isValidElement(child)) {
        return child
      }
      if (!child.props.name && child.props.type !== 'submit') {
        return React.cloneElement(child, {}, this._walkChildren(child.props.children, serverErrors, validatorErrors, foundError))
      }

      const name = child.props.name
      const error = name && errors[name]
      let firstError = false
      if (!foundError && error) {
        foundError = true
        firstError = true
      }

      const newProps = {
        value: this.state.values[name],
        disabled: this.props.working || child.props.type === 'submit' && _.some(validatorErrors),
      }
      if (child.type !== 'button') {
        _.extend(newProps, {
          onModify: (value) => {
            this.onFieldModify(name, value)
            if (child.props.onModify) {
              child.props.onModify(value)
            }
          },
          onValidate: () => this.onFieldValidate(name),
          error: !!error,
          isFirstError: firstError,
          message: error,
        })
      }

      return React.cloneElement(child, newProps, this._walkChildren(child.props.children, serverErrors, validatorErrors, foundError))
    })
  },

  render() {
    return (
      <form className={classNames('fields', this.props.className)} noValidate onSubmit={this.onSubmit}>
        {this._walkChildren(this.props.children, this.props.errors, this.state.errors)}
      </form>
    )
  },
})
