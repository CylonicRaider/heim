/* eslint-disable react/no-multi-comp */

import _ from 'lodash'
import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import Reflux from 'reflux'
import Immutable from 'immutable'

import Embed from '../ui/Embed'
import MessageText from '../ui/MessageText'


const TVActions = Reflux.createActions([
  'changeVideo',
  'changeNotice',
])
_.extend(module.exports, TVActions)

export const TVStore = Reflux.createStore({
  listenables: [
    TVActions,
    {chatChange: Heim.chat.store},
  ],

  init() {
    this.state = Immutable.fromJS({
      video: {
        time: 0,
        messageId: null,
        youtubeId: null,
        youtubeTime: 0,
        title: '',
      },
      notice: {
        time: 0,
        content: '',
      },
    })
  },

  getInitialState() {
    return this.state
  },

  changeVideo(video) {
    this.state = this.state.set('video', Immutable.fromJS(video))
    this.trigger(this.state)
  },

  changeNotice(notice) {
    this.state = this.state.set('notice', Immutable.fromJS(notice))
    this.trigger(this.state)
  },
})

export const SyncedEmbed = createReactClass({
  displayName: 'SyncedEmbed',

  propTypes: {
    messageId: PropTypes.string,
    youtubeId: PropTypes.string,
    youtubeTime: PropTypes.number,
    startedAt: PropTypes.number,
    className: PropTypes.string,
    clientTimeOffset: PropTypes.any,
  },

  shouldComponentUpdate(nextProps) {
    return nextProps.messageId !== this.props.messageId
  },

  render() {
    const timeOffset = this.props.clientTimeOffset ? this.props.clientTimeOffset() : 0
    return (
      <Embed
        className={this.props.className}
        kind="youtube"
        autoplay="1"
        start={Math.max(0, Math.floor(Date.now() / 1000 - this.props.startedAt - timeOffset)) + this.props.youtubeTime}
        youtube_id={this.props.youtubeId}
        messageId={this.props.messageId}
      />
    )
  },
})

export const YouTubeTV = createReactClass({
  displayName: 'YouTubeTV',

  propTypes: {
    clientTimeOffset: PropTypes.any,
  },

  mixins: [
    Reflux.connect(TVStore, 'tv'),
    require('react-addons-pure-render-mixin'),
  ],

  render() {
    return (
      <SyncedEmbed
        className="youtube-tv"
        messageId={this.state.tv.getIn(['video', 'messageId'])}
        youtubeId={this.state.tv.getIn(['video', 'youtubeId'])}
        startedAt={this.state.tv.getIn(['video', 'time'])}
        youtubeTime={this.state.tv.getIn(['video', 'youtubeTime'])}
        clientTimeOffset={this.props.clientTimeOffset}
      />
    )
  },
})

export const YouTubePane = createReactClass({
  displayName: 'YouTubePane',

  propTypes: {
    clientTimeOffset: PropTypes.any,
  },

  mixins: [
    Reflux.connect(TVStore, 'tv'),
    require('react-addons-pure-render-mixin'),
  ],

  render() {
    return (
      <div className="chat-pane-container youtube-pane">
        <div className="top-bar">
          <MessageText className="title" content={':notes: :tv: :notes: ' + this.state.tv.getIn(['video', 'title'])} />
        </div>
        <div className="aspect-wrapper">
          <YouTubeTV clientTimeOffset={this.props.clientTimeOffset} />
        </div>
        <MessageText className="notice-board" content={this.state.tv.getIn(['notice', 'content'])} />
      </div>
    )
  },
})

function parseYoutubeTime(time) {
  const timeReg = /([0-9]+h)?([0-9]+m)?([0-9]+s?)?/
  const match = time.match(timeReg)
  if (!match) {
    return 0
  }
  const hours = parseInt(match[1] || 0, 10)
  const minutes = parseInt(match[2] || 0, 10)
  const seconds = parseInt(match[3] || 0, 10)
  return hours * 3600 + minutes * 60 + seconds
}

export function install() {
  let clientTimeOffset = 0
  Heim.chat.store.socket.on('receive', (ev) => {
    if (ev.type === 'ping-event') {
      clientTimeOffset = Date.now() / 1000 - ev.data.time
    }
  })

  Heim.ui.createCustomPane('youtube-tv', {readOnly: true})

  Heim.hook('thread-panes', () => <YouTubePane key="youtube-tv" clientTimeOffset={() => clientTimeOffset} />)

  Heim.hook('main-pane-top', function YouTubeTVInject() {
    return this.state.ui.thin ? <YouTubeTV key="youtube-tv" clientTimeOffset={() => clientTimeOffset} /> : null
  })

  Heim.chat.messagesChanged.listen((ids, state) => {
    const candidates = Immutable.Seq(ids)
      .map((messageId) => {
        const msg = state.messages.get(messageId)
        const valid = messageId !== '__root' && msg.get('content')
        return valid && msg
      })
      .filter(Boolean)

    const playRe = /!play [^?]*\?v=([-\w]+)(?:&t=([0-9hms]+))?/
    const video = candidates
      .map((msg) => {
        const match = msg.get('content').match(playRe)
        return match && {
          time: msg.get('time'),
          messageId: msg.get('id'),
          youtubeId: match[1],
          youtubeTime: match[2] ? parseYoutubeTime(match[2]) : 0,
          title: msg.get('content'),
        }
      })
      .filter(Boolean)
      .sortBy(v => v.time)
      .last()

    if (video && video.time > TVStore.state.getIn(['video', 'time'])) {
      TVActions.changeVideo(video)
    }

    const noticeRe = /^!notice(\S*?)\s([^]*)$/
    const notices = candidates
      .map((msg) => {
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

    const noticeMaxSummaryLength = 80
    notices.forEach((notice) => {
      const lines = notice.content.split('\n')
      let content = lines[0]
      if (content.length >= noticeMaxSummaryLength || lines.length > 1) {
        content = content.substr(0, noticeMaxSummaryLength) + '…'
      }
      state.messages.mergeNodes(notice.id, {
        content: '/me changed the notice to: "' + content + '"',
      })
    })

    const latestNotice = notices
      .filter(n => n.display)
      .sortBy(notice => notice.time)
      .last()

    if (latestNotice && latestNotice.time > TVStore.state.getIn(['notice', 'time'])) {
      TVActions.changeNotice(latestNotice)
    }
  })

  Heim.hook('page-bottom', () => (
    <link key="youtubetv-style" rel="stylesheet" type="text/css" href="/static/youtube-tv.css" />
  ))
}
