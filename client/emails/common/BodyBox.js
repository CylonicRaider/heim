import React from 'react'
import PropTypes from 'prop-types'

import { Item, Box } from 'react-html-email'

export default function BodyBox(props) {
  return (
    <Item style={{paddingTop: '12px'}}>
      <Box cellPadding={20} width="100%" bgcolor="white" style={{borderBottom: '3px solid #ccc'}}>
        {props.children}
      </Box>
    </Item>
  )
}

BodyBox.propTypes = {
  children: PropTypes.node,
}
