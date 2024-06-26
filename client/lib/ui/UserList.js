import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import Immutable from 'immutable'
import classNames from 'classnames'

import chat from '../stores/chat'
import ui from '../stores/ui'
import forwardProps from '../util/forwardProps'
import MessageText from './MessageText'

function pluralize(count, term, plural = term + 's') {
  return count + ' ' + (count === 1 ? term : plural)
}

export default createReactClass({
  displayName: 'UserList',

  propTypes: {
    users: PropTypes.instanceOf(Immutable.Map),
    selected: PropTypes.instanceOf(Immutable.Set),
  },

  mixins: [require('react-immutable-render-mixin')],

  onMouseDown(ev, sessionId) {
    if (ui.store.state.managerMode) {
      const selected = this.props.selected.has(sessionId)
      chat.setUserSelected(sessionId, !selected)
      ui.startToolboxSelectionDrag(!selected)
      ev.preventDefault()
    }
  },

  onMouseEnter(sessionId) {
    if (ui.store.state.managerMode && ui.store.state.draggingToolboxSelection) {
      chat.setUserSelected(sessionId, ui.store.state.draggingToolboxSelectionToggle)
    }
  },

  render() {
    const list = this.props.users
      .toSeq()
      .filter((user) => user.get('present') && user.get('name'))
      .sortBy((user) => user.get('name').toLowerCase())
      .groupBy((user) => /^bot:/.test(user.get('id')) ? 'bot' : 'human')
    const humanCount = (list.get('human') || {size: 0}).size
    const botCount = (list.get('bot') || {size: 0}).size

    const lurkers = this.props.users
      .toSeq()
      .filter((user) => user.get('present') && !user.get('name'))
      .groupBy((user) => /^bot:/.test(user.get('id')) ? 'bot' : 'human')
    const humanLurkerCount = (lurkers.get('human') || {size: 0}).size
    const botLurkerCount = (lurkers.get('bot') || {size: 0}).size

    const formatCounter = (cntU, cntD, cntL, people) => {
      const tooltipParts = []
      if (cntU) tooltipParts.push(people ? pluralize(cntU, 'person', 'people') : pluralize(cntU, 'bot'))
      if (cntD) tooltipParts.push(people ? pluralize(cntD, 'duplicate person', 'duplicate people') : pluralize(cntD, 'duplicate bot'))
      if (cntL) tooltipParts.push(people ? pluralize(cntL, 'lurker') : pluralize(cntL, 'bot lurker'))
      if (!tooltipParts) tooltipParts.push('(no-one?!)')
      return <span className="user-counter" title={tooltipParts.join('; ')}>({cntU}{cntD ? '(+' + cntD + ')' : ''}{cntL ? '+' + cntL : ''})</span>
    }
    const formatUser = (user) => {
      const sessionId = user.get('session_id')
      const selected = this.props.selected.has(sessionId)
      return (
        <span
          key={sessionId}
          onMouseDown={(ev) => this.onMouseDown(ev, sessionId)}
          onMouseEnter={() => this.onMouseEnter(sessionId)}
        >
          <MessageText
            className={classNames('nick', {'selected': selected})}
            onlyEmoji
            style={{background: 'hsl(' + user.get('hue') + ', 65%, 85%)'}}
            content={user.get('name')}
            title={user.get('name')}
          />
        </span>
      )
    }

    let prevUser
    let people = list.get('human')
    people = people && people.filter((user) => {
      if (prevUser && user.get('id') === prevUser.get('id') && user.get('name') === prevUser.get('name')) {
        return false
      }
      prevUser = user
      return true
    }).toList()
    const humanUniqueCount = (people || {size: 0}).size
    const humanDuplicateCount = humanCount - humanUniqueCount
    const bots = list.get('bot')
    return (
      <div className="user-list" {...forwardProps(this)}>
        {people && (
          <div className="list">
            <h1>people {formatCounter(humanUniqueCount, humanDuplicateCount, humanLurkerCount, true)}</h1>
            {people.map(formatUser).toIndexedSeq()}
          </div>
        )}
        {bots && (
          <div className="list">
            <h1>bots {formatCounter(botCount, 0, botLurkerCount, false)}</h1>
            {bots.map(formatUser).toIndexedSeq()}
          </div>
        )}
      </div>
    )
  },
})
