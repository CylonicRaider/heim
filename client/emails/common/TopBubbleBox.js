import React from 'react'
import PropTypes from 'prop-types'

import { Box, Item, A, Image } from 'react-html-email'

export default function TopBubbleBox(props) {
  return (
    <Item align="center">
      <A href="{{.SiteURL}}">
        <Image src={'{{.File `' + props.logo + '`}}'} alt="Logo" width={67} height={90} />
      </A>
      <Box width="600" cellPadding={2} bgcolor="white" style={{
        borderBottom: '3px solid #ccc',
        borderRadius: '10px',
        padding: props.padding,
      }}>
        {props.children}
      </Box>
    </Item>
  )
}

TopBubbleBox.propTypes = {
  logo: PropTypes.string.isRequired,
  padding: PropTypes.number,
  children: PropTypes.node,
}

TopBubbleBox.defaultProps = {
  padding: 7,
}
