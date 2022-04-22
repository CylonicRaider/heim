import React from 'react'
import PropTypes from 'prop-types'

import MessageText from '../../lib/ui/MessageText'
import hueHash from '../../lib/hueHash'

export default function FauxNick(props) {
  return <MessageText className="nick" onlyEmoji style={{background: 'hsl(' + hueHash.hue(props.nick) + ', 65%, 85%)'}} content={props.nick} />
}

FauxNick.propTypes = {
  nick: PropTypes.string,
}
