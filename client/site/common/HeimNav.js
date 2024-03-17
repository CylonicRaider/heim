import React from 'react'
import PropTypes from 'prop-types'

import NavBar from './NavBar'

const items = [
  {name: 'opensource', caption: <span><span className="long">Open </span>Source</span>},
]

export default function HeimNav(props) {
  return <NavBar prefix="/heim" items={items} {...props} />
}

HeimNav.propTypes = {
  selected: PropTypes.string,
}
