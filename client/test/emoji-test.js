import './support/setup'
import assert from 'assert'

describe('emoji', () => {
  const emoji = require('../lib/emoji')

  describe('unicodeToIconID', () => {
    it('translates regular emoji', () => {
      // U+25B6 BLACK RIGHT-POINTING TRIANGLE (:arrow_forward:)
      assert.equal(emoji.unicodeToIconID('\u25b6'), '25b6')
    })

    it('translates a non-BMP emoji', () => {
      // U+1F514 BELL (:bell:)
      assert.equal(emoji.unicodeToIconID('\ud83d\udd14'), '1f514')
    })
  })
})
