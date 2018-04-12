import React from 'react'
import PropTypes from 'prop-types'


export default function HeimAttachPoint(props) {
  return <div id={props.id} data-context="{{.Data}}" />
}

HeimAttachPoint.propTypes = {
  id: PropTypes.string,
}
