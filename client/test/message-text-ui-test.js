import assert from 'assert'
import React from 'react'
import ReactDOM from 'react-dom'
import ReactTestUtils from 'react-dom/test-utils'
import support from './support/setup'

import MessageText from '../lib/ui/MessageText'


describe('<MessageText>', () => {
  support.fakeEnv({
    HEIM_PREFIX: '/test',
  })

  function renderMessageText(content) {
    return ReactTestUtils.renderIntoDocument(
      <MessageText content={content} />
    )
  }

  it('automatically links urls', () => {
    const messageContent = renderMessageText('http://google.com')
    assert.equal(ReactDOM.findDOMNode(messageContent).innerHTML,
      '<a href="http://google.com" target="_blank" rel="noreferrer">google.com</a>')
  })

  it('truncates long urls', () => {
    const messageContent = renderMessageText('http://google.com/abcdefghijklmnopqrstuvwxyz1234567890')
    assert.equal(ReactDOM.findDOMNode(messageContent).innerHTML,
      '<a href="http://google.com/abcdefghijklmnopqrstuvwxyz1234567890" target="_blank" rel="noreferrer" title="http://google.com/abcdefghijklmnopqrstuvwxyz1234567890">google.com/abcdefghijklmnopqrstuvwxyz\u2026</a>')
  })

  it('linkifies &room references', () => {
    const messageContent = renderMessageText('hello &space! foo&bar &bar &baz')
    assert.equal(ReactDOM.findDOMNode(messageContent).innerHTML,
      'hello <a href="/test/room/space/" target="_blank">&amp;space</a>! foo&amp;bar <a href="/test/room/bar/" target="_blank">&amp;bar</a> <a href="/test/room/baz/" target="_blank">&amp;baz</a>')
  })

  it('doesn\'t linkify javascript:// links', () => {
    const messageContent = renderMessageText('Javascript://hello javascript://world')  // eslint-disable-line no-script-url
    assert.equal(ReactDOM.findDOMNode(messageContent).innerHTML,
      'Javascript://hello javascript://world')  // eslint-disable-line no-script-url
  })

  it('processes emoji', () => {
    // Unicode / emoji cheat sheet:
    // U+25B6 BLACK RIGHT-POINTING TRIANGLE (:arrow_forward:)
    // U+1F514 BELL (:bell:)
    // U+2122 TRADE MARK SIGN (:tm:)
    // U+00A9 COPYRIGHT SIGN (no emoji)
    const messageContent = renderMessageText(':euphoria: \u25b6 \ud83d\udd14 \u2122 \u00a9')
    assert.equal(ReactDOM.findDOMNode(messageContent).innerHTML,
      '<span class="emoji emoji-euphoria" title=":euphoria:">:euphoria:</span> <span class="emoji emoji-25b6" title=":arrow_forward:">\u25b6</span> <span class="emoji emoji-1f514" title=":bell:">\ud83d\udd14</span> \u2122 \u00a9')
  })
})
