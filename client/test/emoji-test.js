import './support/setup'
import assert from 'assert'

describe('emoji', () => {
  const emoji = require('../lib/emoji')

  describe('nameToUnicode', () => {
    it('translates BMP emoji', () => {
      // U+25C0 BLACK LEFT-POINTING TRIANGLE (:arrow_backward:)
      assert.equal(emoji.nameToUnicode('arrow_backward'), '\u25c0')
    })

    it('translates astral emoji', () => {
      // U+1F34E RED APPLE (:apple:)
      assert.equal(emoji.nameToUnicode('apple'), '\ud83c\udf4e')
    })

    it('does not translate custom emoji', () => {
      assert.equal(emoji.nameToUnicode('euphoria'), null)
    })
  })

  describe('unicodeToIconID', () => {
    it('translates BMP emoji', () => {
      // U+25B6 BLACK RIGHT-POINTING TRIANGLE (:arrow_forward:)
      assert.equal(emoji.unicodeToIconID('\u25b6'), '25b6')
    })

    it('translates astral emoji', () => {
      // U+1F514 BELL (:bell:)
      assert.equal(emoji.unicodeToIconID('\ud83d\udd14'), '1f514')
    })
  })
})
