import './support/setup'
import assert from 'assert'

import emoji from '../lib/heim/emoji'

const referenceEmoji = [
  { u: '\u25b6', i: '25b6', n: 'arrow_forward', d: 'BMP emoji' },
  { u: '\ud83d\udd14', i: '1f514', n: 'bell', d: 'astral emoji' },
  // The Unicode text is incorrect (note the missing \u200d), but this captures the current behavior.
  { u: '\ud83d\udc41\ud83d\udde8', i: '1f441-1f5e8', n: 'eye_speech_bubble', d: 'ZWJ sequences' },
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
