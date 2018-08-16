import _ from 'lodash'
import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import ReactDOMServer from 'react-dom/server'
import Autolinker from 'autolinker'
import twemoji from 'twemoji'
import emoji from '../emoji'

import chat from '../stores/chat'
import hueHash from '../hueHash'
import heimURL from '../heimURL'


const autolinker = new Autolinker({
  truncate: 40,
  replaceFn(match) {
    if (match.getType() === 'url') {
      const url = match.getUrl()
      const tag = this.getTagBuilder().build(match)

      if (/^javascript/.test(url.toLowerCase())) {
        // Thanks, Jordan!
        return false
      }

      if (location.protocol === 'https:' && RegExp('^https?://' + location.hostname).test(url)) {
        // self-link securely
        tag.setAttr('href', url.replace(/^http:/, 'https:'))
      } else {
        tag.setAttr('rel', 'noreferrer')
      }

      return tag
    }
    return null
  },
})

export default createReactClass({
  displayName: 'MessageText',

  propTypes: {
    content: PropTypes.string.isRequired,
    maxLength: PropTypes.number,
    onlyEmoji: PropTypes.bool,
    className: PropTypes.string,
    title: PropTypes.string,
    style: PropTypes.object,
  },

  mixins: [
    require('react-immutable-render-mixin'),
  ],

  render() {
    /* eslint-disable react/no-danger */
    // FIXME: replace with React splitting parser + preserve links when trimmed

    let content = this.props.content

    if (this.props.maxLength) {
      content = _.truncate(content, {length: this.props.maxLength})
    }

    let html = _.escape(content)

    if (!this.props.onlyEmoji) {
      html = html.replace(/\B&amp;(\w+)(?=$|[^\w;])/g, (match, name) =>
        ReactDOMServer.renderToStaticMarkup(<a href={heimURL('/room/' + name + '/')} target="_blank">&amp;{name}</a>)
      )

      html = html.replace(chat.mentionFindRe, (match, pre, name) => {
        const color = 'hsl(' + hueHash.hue(name) + ', 50%, 42%)'
        return pre + ReactDOMServer.renderToStaticMarkup(<span style={{color: color}} className="mention-nick">@{name}</span>)
      })
    }

    html = html.replace(emoji.namesRe, (match, name) =>
      ReactDOMServer.renderToStaticMarkup(<span className={'emoji emoji-' + emoji.index[name]} title={match}>{match}</span>)
    )

    html = twemoji.replace(html, (match) => {
      const codePoint = emoji.lookupEmojiCharacter(match)
      if (!codePoint) {
        return match
      }
      const emojiName = emoji.names[codePoint] ? ':' + emoji.names[codePoint] + ':' : match
      return ReactDOMServer.renderToStaticMarkup(<span className={'emoji emoji-' + codePoint} title={emojiName}>{match}</span>)
    })

    if (!this.props.onlyEmoji) {
      html = autolinker.link(html)
    }

    return (
      <span className={this.props.className} style={this.props.style} title={this.props.title} dangerouslySetInnerHTML={{
        __html: html,
      }} />
    )
  },
})
