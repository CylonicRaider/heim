import React from 'react'
import PropTypes from 'prop-types'

import { Email } from 'react-html-email'


export default function StandardEmail(props) {
  return (
    <Email title="{{.Subject}}" bgcolor="#f0f0f0" cellSpacing={10} style={{paddingTop: '20px', paddingBottom: '20px'}}>
      {props.children}
    </Email>
  )
}

StandardEmail.propTypes = {
  children: PropTypes.node,
}
