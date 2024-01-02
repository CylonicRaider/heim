import React from 'react'

import heimURL from '../../lib/heimURL'

export default function Footer() {
  return (
    <footer>
      <div className="container">
        <a href={heimURL('/about/terms')}>terms<span className="long"> of service</span></a>
        <a href={heimURL('/about/privacy')}>privacy<span className="long"> policy</span></a>
        <span className="spacer" />
        <a href={heimURL('/about')}>about</a>
        <a href={heimURL('/about/values')}>values</a>
        <a href={heimURL('/about/conduct')}><span className="long">code of </span>conduct</a>
        <span className="spacer" />
        <a href="https://github.com/CylonicRaider/heim"><span className="long">source </span>code</a>
      </div>
    </footer>
  )
}
