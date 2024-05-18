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
  runTests('nameToIconID', (record) => {
    assert.equal(emoji.nameToIconID(record.n), record.i)
  })

  runTests('iconIDToName', (record) => {
    assert.equal(emoji.iconIDToName(record.i), record.n)
  })

  runTests('iconClass', (record) => {
    assert.ok(/^emoji emoji-[a-z0-9-]+$/.test(emoji.iconClass(record.i)))
  })

  runTests('nameToUnicode', (record) => {
    assert.equal(emoji.nameToUnicode(record.n), record.u)
  })

  runTests('unicodeToIconID', (record) => {
    assert.equal(emoji.unicodeToIconID(record.u), record.i)
  }, (record) => record.u)
})
