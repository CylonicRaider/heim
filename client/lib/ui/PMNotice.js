import React from 'react'
import PropTypes from 'prop-types'

import chat from '../stores/chat'
import ui from '../stores/ui'
import hueHash from '../hueHash'
import FastButton from './FastButton'


export default function PMNotice(props) {
  const bgColor = 'hsl(' + hueHash.hue(props.nick) + ', 67%, 85%)'
  const textLightColor = 'hsl(' + hueHash.hue(props.nick) + ', 28%, 28%)'
  const textColor = 'hsl(' + hueHash.hue(props.nick) + ', 28%, 43%)'
  return (
    <div className="notice light pm" style={{backgroundColor: bgColor}}>
      <div className="content">
        <span className="title" style={{color: textLightColor}}>{props.kind === 'from' ? `${props.nick} invites you to a private conversation` : `you invited ${props.nick} to a private conversation`}</span>
        <div className="actions">
          <FastButton onClick={() => ui.openPMWindow(props.pmId)} style={{color: textColor}}>join room</FastButton>
        </div>
      </div>
      <FastButton className="close" onClick={() => chat.dismissPMNotice(props.pmId)} />
    </div>
  )
}

PMNotice.propTypes = {
  pmId: PropTypes.string.isRequired,
  nick: PropTypes.string.isRequired,
  kind: PropTypes.string.isRequired,
}
