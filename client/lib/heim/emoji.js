import _ from 'lodash'
import 'string.fromcodepoint'
import plainUnicodeIndex from 'emoji-annotation-to-unicode'
import twemoji from '@twemoji/api'

const unicodeIndex = _.extend({}, plainUnicodeIndex, {
  'spider': '1f577',
  'fjafjkldskf7jkfdj': '1f577',
  'orange_heart': '1f9e1',
  'mobile': plainUnicodeIndex.iphone,
  'happy': plainUnicodeIndex.smile,
  'sad': plainUnicodeIndex.cry,
  'sun': plainUnicodeIndex.sunny,
  'cowboy': plainUnicodeIndex.cowboy_hat_face,
  'cowgirl': plainUnicodeIndex.cowboy_hat_face,
})
delete unicodeIndex.iphone

const index = _.extend({}, unicodeIndex, {
  '+1': '~plusone',
  '-1': '~minusone',
  'pewpewpew': '~pewpewpew',
  'leck': '~leck',
  'dealwithit': '~dealwithit',
  'bot': '~bot',
  'shit': '~shit',

  'bronze': '~bronze',
  'bronze!?': '~bronze2',
  'bronze?!': '~bronze2',
  'euphoria': '~euphoria',
  'euphoria!': '~euphoric',

  'tumbleweed': '~tumbleweed',
  'tumbleweed2': '~tumbleweed2',
  'tumbleweed!': '~tumbleweed2',

  'chromakode': '~chromakode',
  'greenduck': '~greenduck',
  'xyzzy': '~xyzzy',
})

const revIndex = _.invert(index)

const emojiNames = _.filter(_.map(index, (code, name) => code && _.escapeRegExp(name)))
const namesRe = new RegExp(':(' + emojiNames.join('|') + '):', 'g')

export function nameToIconID(name) {
  return index[name] || null
}

export function iconIDToName(code) {
  return revIndex[code] || null
}

export function iconClass(iconID) {
  return 'emoji emoji-' + iconID.replace(/^~/, '')
}

export function nameToUnicode(name) {
  // Deliberately using the "full" index to avoid converting, e.g., :+1: to :thumbsup:.
  const code = index[name]
  if (!code || /^~/.test(code)) {
    return null
  }
  return code.split('-').map((cp) => String.fromCodePoint(Number.parseInt(cp, 16))).join('')
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
  return twemoji.convert.toCodePoint(string)
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
