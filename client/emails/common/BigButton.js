import React from 'react'
import PropTypes from 'prop-types'

import { Item, A, Span, textDefaults } from 'react-html-email'

export default function BigButton(props) {
  return (
    <Item align="center">
      <A color="white" textDecoration="none" href={props.href} style={{
        background: props.color,
        padding: '22px 30px',
        borderRadius: '4px',
      }}>
        <Span {...textDefaults} fontSize={24} fontWeight="bold" color="white">{props.children}</Span>
      </A>
    </Item>
  )
}

BigButton.propTypes = {
  href: PropTypes.string.isRequired,
  color: PropTypes.string,
  children: PropTypes.node,
}
