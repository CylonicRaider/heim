import React from 'react'
import PropTypes from 'prop-types'

import heimURL from '../../lib/heim/heimURL'

export default function Page(props) {
  return (
    <html lang="en-US">
      <head>
        <meta charSet="utf-8" />
        <title>{props.title}</title>
        <link rel="icon" id="favicon" href={heimURL('/static/favicon.png')} sizes="32x32" />
        <link rel="icon" href={heimURL('/static/favicon-192.png')} sizes="192x192" />
        <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no" />
        {props.meta || null}
        {props.heimPage && <link rel="stylesheet" type="text/css" id="css" href={heimURL('/static/main.css')} />}
        <link rel="stylesheet" type="text/css" id="css" href={heimURL('/static/site.css')} />
        {props.heimPage && <script src={heimURL('/static/raven.js')} />}
        <script defer src={heimURL('/static/site.js')} />
      </head>
      <body className={props.className}>
        {props.children}
        {props.heimPage && <script defer id="heim-js" src={heimURL('/static/main.js')} data-entrypoint={props.heimPage} />}
      </body>
    </html>
  )
}

Page.propTypes = {
  title: PropTypes.string,
  meta: PropTypes.node,
  heimPage: PropTypes.string,
  className: PropTypes.string,
  children: PropTypes.node,
}
