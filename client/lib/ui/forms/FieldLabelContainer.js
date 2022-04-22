import React from 'react'
import PropTypes from 'prop-types'
import classNames from 'classnames'

export default function FieldLabelContainer(props) {
  /* eslint-disable jsx-a11y/label-has-associated-control */
  return (
    <label className={classNames('field-label-container', props.error && 'error', props.className)}>
      <div className="label-line">
        <div className="label">{props.label}</div>
        {props.action && <button type="button" className="action" onClick={props.onAction}>{props.action}</button>}
        <div className="spacer" />
        {props.message && <div className="message">{props.message}</div>}
      </div>
      {props.children}
    </label>
  )
}

FieldLabelContainer.propTypes = {
  label: PropTypes.string.isRequired,
  className: PropTypes.string,
  action: PropTypes.string,
  onAction: PropTypes.func,
  error: PropTypes.bool,
  message: PropTypes.string,
  children: PropTypes.node,
}
