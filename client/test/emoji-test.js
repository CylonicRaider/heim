import './support/setup'
import assert from 'assert'

describe('emoji', () => {
  const emoji = require('../lib/emoji')

  describe('toCodePoint', () => {
    it('translates regular emoji', () => {
      // U+25B6 BLACK RIGHT-POINTING TRIANGLE (:arrow_forward:)
      assert.equal(emoji.lookupEmojiCharacter('\u25b6'), '25b6')
    })

    it('translates a non-BMP emoji', () => {
      // U+1F514 BELL (:bell:)
      assert.equal(emoji.lookupEmojiCharacter('\ud83d\udd14'), '1f514')
    })
  })
})
