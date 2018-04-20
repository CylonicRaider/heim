/* eslint-disable react/no-danger */

import React from 'react'
import moment from 'moment'


export default function initPlugins(roomName) {
  if (roomName === 'thedrawingroom' || roomName === 'lovenest' || roomName === 'has') {
    Heim.hook('page-bottom', () => (
      <link key="thedrawingroom-style" rel="stylesheet" type="text/css" href="/static/thedrawingroom.css" />
    ))
  }

  if (roomName === 'space') {
    const Embed = require('./ui/Embed').default

    Heim.hook('main-sidebar', () => (
      <div key="norman" className="norman">
        <p>norman</p>
        <Embed kind="imgur" imgur_id="UKbitCO" />
      </div>
    ))

    Heim.hook('page-bottom', () => (
      <link key="norman-style" rel="stylesheet" type="text/css" href="/static/norman.css" />
    ))
  }

  if (roomName === 'music' || roomName === 'youtube' || roomName === 'rmusic' || roomName === 'listentothis') {
    require('./gadgets/YoutubeTV').install()
  }

  if (roomName === 'adventure' || roomName === 'chess' || roomName === 'monospace') {
    Heim.hook('page-bottom', () => (
      <link key="monospace-style" rel="stylesheet" type="text/css" href="/static/monospace.css" />
    ))

    Heim.chat.setRoomSettings({collapse: false})
  }

  if (uiwindow.location.hash.substr(1) === 'spooky') {
    Heim.hook('page-bottom', () => (
      <link key="spooky-theme" rel="stylesheet" type="text/css" href="/static/theme-spooky.css" />
    ))
  }

  if (roomName === 'sandersforpresident') {
    Heim.hook('main-pane-top', () => {
      const MessageText = require('./ui/MessageText').default
      return (
        <div key="sanders-top-bar" className="secondary-top-bar"><MessageText onlyEmoji content=":us:" /> Welcome to the <a href="https://reddit.com/r/sandersforpresident" target="_blank" rel="noreferrer noopener">/r/SandersForPresident</a> live chat! Please <a href="https://www.reddit.com/r/SandersForPresident/wiki/livechat" target="_blank" rel="noreferrer noopener">read our rules</a>.</div>
      )
    })

    Heim.hook('page-bottom', () => (
      <link key="sanders-style" rel="stylesheet" type="text/css" href="/static/sandersforpresident.css" />
    ))
  }

  if (uiwindow.location.hash.substr(1) === 'darcula') {
    Heim.hook('page-bottom', () => (
      <link key="darcula-theme" rel="stylesheet" type="text/css" href="/static/theme-darcula.css" />
    ))
  }

  if (roomName === 'xkcd') {
    Heim.hook('main-pane-top', () => (
      <div key="xkcd-top-bar" className="secondary-top-bar"><span className="motto" title="All problems are solvable by being thrown at with bots">Omnes qu&aelig;stiones solvuntur eis iactandis per machinis</span></div>
    ))

    Heim.hook('page-bottom', () => (
      <link key="xkcd-style" rel="stylesheet" type="text/css" href="/static/xkcd.css" />
    ))

    if (uiwindow.location.hash.substr(1) === 'spooky') {
      Heim.hook('page-bottom', () => (
        <style key="xkcd-spooky-style" dangerouslySetInnerHTML={{__html: `
            .secondary-top-bar {
              color: darkorange;
              background: #2e293c;
            }
          `}} />
      ))
    } else if (uiwindow.location.hash.substr(1) === 'darcula') {
      Heim.hook('page-bottom', () => (
        <style key="xkcd-darcula-style" dangerouslySetInnerHTML={{__html: `
            .secondary-top-bar {
              color: #758076;
              background: #4c5053;
            }
          `}} />
      ))
    }
  }

  const now = moment()
  if (now.month() === 11 && (now.date() === 13 || now.date() === 14)) {
    Heim.hook('page-bottom', () => (
      <link key="anniversary-style" rel="stylesheet" type="text/css" href="/static/anniversary.css" />
    ))
  }
}
