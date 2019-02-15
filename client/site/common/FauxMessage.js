import React from 'react'
import PropTypes from 'prop-types'

import MessageText from '../../lib/ui/MessageText'
import FauxNick from './FauxNick'


export default function FauxMessage(props) {
  return (
    <div className="faux-message">
      <div className="line">
        <FauxNick nick={props.sender} />
        <div className="content">
          <MessageText className="message" content={props.message} />
          {props.embed && (
            <div className="embed">
              <div className="wrapper">
                <img className="embed" src={props.embed} alt="" />
              </div>
            </div>
          )}
        </div>
      </div>
      {props.children}
    </div>
  )
}

FauxMessage.propTypes = {
  sender: PropTypes.string,
  message: PropTypes.string,
  embed: PropTypes.string,
  children: PropTypes.node,
}
