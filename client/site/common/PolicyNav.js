import React from 'react'
import PropTypes from 'prop-types'

import NavBar from './NavBar'

const items = [
  {name: 'values', caption: <span>Values</span>},
  {name: 'conduct', caption: <span><span className="long">Code of </span>Conduct</span>},
  {name: 'hosts', caption: <span><span className="long">Hosting </span>Rooms</span>},
  {name: 'terms', caption: <span>Terms<span className="long"> of Service</span></span>},
  {name: 'privacy', caption: <span>Privacy<span className="long"> Policy</span></span>},
]

export default function PolicyNav(props) {
  return <NavBar prefix="/about" items={items} {...props} />
}

PolicyNav.propTypes = {
  selected: PropTypes.string,
}
