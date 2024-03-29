import _ from 'lodash'
import 'string.fromcodepoint'
import unicodeIndex from 'emoji-annotation-to-unicode'
import twemoji from 'twemoji'

const index = _.extend({}, unicodeIndex, {
  '+1': 'plusone',
  'bronze': 'bronze',
  'bronze!?': 'bronze2',
  'bronze?!': 'bronze2',
  'euphoria': 'euphoria',
  'euphoria!': 'euphoric',
  'chromakode': 'chromakode',
  'pewpewpew': 'pewpewpew',
  'leck': 'leck',
  'dealwithit': 'dealwithit',
  'spider': 'spider',
  'indigo_heart': 'indigo_heart',
  'orange_heart': 'orange_heart',
  'bot': 'bot',
  'greenduck': 'greenduck',
  'mobile': unicodeIndex.iphone,
  'poop': 'poop',
  'sad': unicodeIndex.cry,
  'sun': unicodeIndex.sunny,
  'cowboy': unicodeIndex.cowboy_hat_face,
  'cowgirl': unicodeIndex.cowboy_hat_face,
  'tumbleweed': 'tumbleweed',
  'tumbleweed2': 'tumbleweed2',
  'tumbleweed!': 'tumbleweed2',
  'fjafjkldskf7jkfdj': 'spider',
})

delete index.iphone

const names = _.invert(index)
const codes = _.uniq(_.values(index))

const emojiNames = _.filter(_.map(index, (code, name) => code && _.escapeRegExp(name)))
const namesRe = new RegExp(':(' + emojiNames.join('|') + '):', 'g')

function nameToUnicode(name) {
  const code = unicodeIndex[name]
  if (!code) {
    return null
  }
  return String.fromCodePoint(Number.parseInt(code, 16))
}

export function lookupEmojiCharacter(icon) {
  // U+FE0E VARIATION SELECTOR-14 signifies text-style display; bail out.
  if (/\ufe0e/.test(icon)) {
    // FIXME: This might appear in the middle of an emoji sequence; I'm not
    //        entirely sure about the semantics in that case.
    return null
  }
  // Algorithm adapted from grabTheRightIcon() from twemoji.
  if (!/\u200d/.test(icon)) {
    icon = icon.replace(/\ufe0f/g, '')
  }
  return twemoji.convert.toCodePoint(icon)
}

export default { index, names, codes, namesRe, nameToUnicode, lookupEmojiCharacter }
