import './support/setup'
import assert from 'assert'

import emoji from '../lib/heim/emoji'

const referenceEmoji = [
  { u: '\u2705', i: '2705', n: 'white_check_mark', d: 'BMP emoji' },
  { u: '\u25b6\ufe0f', i: '25b6-fe0f', n: 'arrow_forward', d: 'BMP default-text emoji' },
  { u: '\ud83d\udd14', i: '1f514', n: 'bell', d: 'astral emoji' },
  { u: '\ud83d\udc41\ufe0f\u200d\ud83d\udde8\ufe0f', i: '1f441-fe0f-200d-1f5e8-fe0f', n: 'eye_speech_bubble', d: 'ZWJ sequences' },
  { u: null, i: '~plusone', n: '+1', d: 'customized Unicode emoji' },
  { u: null, i: '~euphoria', n: 'euphoria', d: 'custom emoji' },
]

function runTests(label, callback, filter = null) {
  describe(label, () => {
    referenceEmoji.forEach((record) => {
      if (filter && !filter(record)) return
      it(`works with ${record.d}`, () => callback(record))
    })
  })
}

describe('emoji', () => {
  runTests('nameToEmojiID', (record) => {
    assert.equal(emoji.nameToEmojiID(record.n), record.i)
  })

  runTests('emojiIDToName', (record) => {
    assert.equal(emoji.emojiIDToName(record.i), record.n)
  })

  runTests('iconClass', (record) => {
    const iconClass = emoji.iconClass(record.i)
    assert.ok(/^emoji emoji-[a-z0-9-]+$/.test(iconClass))
    assert.ok(!/^\b(200d|fe0f)\b/.test(iconClass))
  })

  runTests('nameToUnicode', (record) => {
    assert.equal(emoji.nameToUnicode(record.n), record.u)
  })

  runTests('unicodeToEmojiID', (record) => {
    assert.equal(emoji.unicodeToEmojiID(record.u), record.i)
  }, (record) => record.u)
})
