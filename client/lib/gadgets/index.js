import React from 'react'
import moment from 'moment'
import queryString from 'querystring'

import heimURL from '../heim/heimURL'

const roomStylesheets = {
  ezziethedog: 'room-grayscale',
  adventure: 'room-monospace',
  monospace: 'room-monospace',
  xkcd: 'room-xkcd',
}

export default function initPlugins(roomName) {
  /* Custom stylesheets */
  // Add a new .less file and amend the table above to create a new one.
  const stylesheet = roomStylesheets[roomName]
  if (stylesheet) {
    Heim.hook('page-bottom', () => (
      <link key="custom-style" rel="stylesheet" type="text/css" href={heimURL('/static/' + stylesheet + '.css')} />
    ))
  }

  /* Per-room customizations */
  if (roomName === 'music' || roomName === 'youtube') {
    require('./YoutubeTV').install()
  }

  if (roomName === 'adventure' || roomName === 'monospace') {
    Heim.chat.setRoomSettings({collapse: false})
  }

  if (roomName === 'xkcd') {
    Heim.hook('main-pane-top', () => (
      <div key="xkcd-top-bar" className="secondary-top-bar"><span className="motto" title="All problems are solvable by being thrown at with bots">Omnes qu&aelig;stiones solvuntur eis iactandis per machinis</span></div>
    ))
  }

  /* Alternate themes */
  const hashFlags = queryString.parse(uiwindow.location.hash.substr(1))
  const ThemeChooser = require('./ThemeChooser')
  ThemeChooser.install({theme: hashFlags.theme})

  /* Anniversary and Halloween gadgets */
  const now = moment()
  if (now.month() === 9 && now.date() === 31) {
    // FIXME: Set the theme only once
    if (!hashFlags.theme) {
      setImmediate(() => ThemeChooser.setTheme('spooky'))
    }
  } else if (now.month() === 11 && (now.date() === 13 || now.date() === 14)) {
    Heim.hook('page-bottom', () => (
      <link key="anniversary-style" rel="stylesheet" type="text/css" href={heimURL('/static/anniversary.css')} />
    ))
  }
}
