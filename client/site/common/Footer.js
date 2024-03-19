import React from 'react'

import heimURL from '../../lib/heimURL'

import links from './links'

export default function Footer() {
  return (
    <footer>
      <div className="container">
        <a href={heimURL('/about/terms')}>Terms<span className="long"> of Service</span></a>
        <a href={heimURL('/about/privacy')}>Privacy<span className="long"> Policy</span></a>
        <span className="spacer" />
        <a href={heimURL('/about')}>About</a>
        <a href={heimURL('/about/values')}>Values</a>
        <a href={heimURL('/about/conduct')}><span className="long">Code of </span>Conduct</a>
        <a href={heimURL('/about/hosts')}>Hosting<span className="long"> Policy</span></a>
        <span className="spacer" />
        <a href={heimURL('/heim/api')}>API</a>
        <a href={heimURL('/heim')}><span className="long">Source </span>Code</a>
      </div>
    </footer>
  )
}
