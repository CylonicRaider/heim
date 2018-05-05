import React from 'react'
import createReactClass from 'create-react-class'
import Reflux from 'reflux'
import Immutable from 'immutable'

import forwardProps from '../forwardProps'
import chat from '../stores/chat'
import MessageText from '../ui/MessageText'


const noticeRe = /^!notice(\S*?)\s([^]*)$/
const noticeMaxSummaryLength = 80

export const store = Reflux.createStore({
  listenables: [
    {messagesChanged: chat.messagesChanged},
  ],

  init() {
    this.state = Immutable.fromJS({time: null, content: ''})
  },

  getInitialState() {
    return this.state
  },

  messagesChanged(ids, state) {
    // Retrieve notice-changing messages
    const notices = Immutable.Seq(ids)
      .map((messageId) => {
        const msg = state.messages.get(messageId)
        if (messageId === '__root' || !msg.get('content')) {
          return false
        }

        const match = msg.get('content').match(noticeRe)
        return match && {
          id: msg.get('id'),
          time: msg.get('time'),
          display: !match[1].length,
          content: match[2],
        }
      })
      .filter(Boolean)
      .cacheResult()

    // Abbreviate the messages
    notices.forEach((notice) => {
      const lines = notice.content.split('\n')
      let content = lines[0]
      if (content.length >= noticeMaxSummaryLength || lines.length > 1) {
        content = content.substr(0, noticeMaxSummaryLength) + '…'
      }
      state.messages.mergeNodes(notice.id, {
        content: '/me changed the notice to: “' + content + '”',
      })
    })

    // Select the newest notice change
    const latestNotice = notices
      .filter(n => n.display)
      .maxBy(n => n.time)

    // Incorporate it into our state
    if (latestNotice && latestNotice.time > this.state.get('time')) {
      this.state = this.state.merge({
        time: latestNotice.time,
        content: latestNotice.content,
      })
      this.trigger(this.state)
    }
  },
})

export const NoticeBoard = createReactClass({
  displayName: 'NoticeBoard',

  mixins: [
    Reflux.connect(store, 'data'),
  ],

  render() {
    return (
      <MessageText {...forwardProps(this)} content={this.state.data.get('content')} />
    )
  },
})
