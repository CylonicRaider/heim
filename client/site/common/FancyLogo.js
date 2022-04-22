import React from 'react'

import heimURL from '../../lib/heimURL'

export default function FancyLogo() {
  return (
    <div className="fancy-logo">
      <a className="logo" href={heimURL('/room/welcome/')} tabIndex={1}>welcome</a>
      <div className="colors">
        <div className="a" />
        <div className="b" />
        <div className="c" />
        <div className="d" />
        <div className="e" />
      </div>
    </div>
  )
}
