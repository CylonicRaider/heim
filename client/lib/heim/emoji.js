import _ from 'lodash'
import 'string.fromcodepoint'
import plainUnicodeIndex from 'emoji-annotation-to-unicode'
import twemoji from 'twemoji'

const unicodeIndex = _.mapValues(plainUnicodeIndex, (n) => 'u/' + n)
_.extend(unicodeIndex, {
  'spider': 'u/1f577',
  'fjafjkldskf7jkfdj': 'u/1f577',
  'orange_heart': 'u/1f9e1',
  'mobile': unicodeIndex.iphone,
  'happy': unicodeIndex.smile,
  'sad': unicodeIndex.cry,
  'sun': unicodeIndex.sunny,
  'cowboy': unicodeIndex.cowboy_hat_face,
  'cowgirl': unicodeIndex.cowboy_hat_face,
})
delete unicodeIndex.iphone

const index = _.extend({}, unicodeIndex, {
  '+1': 'c/plusone',
  'pewpewpew': 'c/pewpewpew',
  'leck': 'c/leck',
  'dealwithit': 'c/dealwithit',
  'bot': 'c/bot',
  'shit': 'c/shit',

  'bronze': 'c/bronze',
  'bronze!?': 'c/bronze2',
  'bronze?!': 'c/bronze2',
  'euphoria': 'c/euphoria',
  'euphoria!': 'c/euphoric',

  'tumbleweed': 'c/tumbleweed',
  'tumbleweed2': 'c/tumbleweed2',
  'tumbleweed!': 'c/tumbleweed2',

  'chromakode': 'c/chromakode',
  'greenduck': 'c/greenduck',
  'xyzzy': 'c/xyzzy',
})

const revIndex = _.invert(index)

const emojiNames = _.filter(_.map(index, (code, name) => code && _.escapeRegExp(name)))
const namesRe = new RegExp(':(' + emojiNames.join('|') + '):', 'g')

export function nameToIconID(name) {
  const code = index[name]
  if (!code || !/^[uc]\//.test(code)) {
    return null
  }
  return code
}

export function iconIDToName(code) {
  return revIndex[code] || null
}

export function iconClass(iconID) {
  return 'emoji emoji-' + iconID.replace(/^[uc]\//, '')
}

export function nameToUnicode(name) {
  const code = unicodeIndex[name]
  if (!code || !/^u\//.test(code)) {
    return null
  }
  return code.substring(2).split('-').map((cp) => String.fromCodePoint(Number.parseInt(cp, 16))).join('')
}

export function unicodeToIconID(string) {
  // U+FE0E VARIATION SELECTOR-14 signifies text-style display; bail out.
  if (/\ufe0e/.test(string)) {
    // FIXME: This might appear in the middle of an emoji sequence; I'm not
    //        entirely sure about the semantics in that case.
    return null
  }
  // Algorithm adapted from grabTheRightIcon() from twemoji.
  if (!/\u200d/.test(string)) {
    string = string.replace(/\ufe0f/g, '')
  }
  // Despite its name, the function handles multi-codepoint emoji correctly
  return 'u/' + twemoji.convert.toCodePoint(string)
}

export default {
  unicodeIndex,
  index,
  revIndex,
  namesRe,
  nameToIconID,
  iconIDToName,
  iconClass,
  nameToUnicode,
  unicodeToIconID,
}
