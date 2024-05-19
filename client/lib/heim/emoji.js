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

// The handling of emoji sequences is a huge mess.
// twemoji includes a function to map an code point sequence to a file name within twemoji, but
// twemoji *also* includes files that are not reachable by that function.
// Additionally, we want our emoji IDs to be mappable to Unicode strings that form correct
// fully-qualified emoji sequences.
// Therefore, we:
// - Clean up code point sequences in emoji-annotation-to-unicode to be fully qualified (coming
//   soon);
// - Normalize code point sequences after they have been isolated by stripping all ZWJs and VS15s
//   (once a code point sequence has been identified as a single emoji, the two characters only
//   create the abovementioned confusion) and looking up the correct fully-qualified form;
// - Strip all ZWJs and VS15s from emoji CSS class names to save a little space.
const codepointIndex = _.keyBy(_.values(unicodeIndex), (code) => code.replace(/-(200d|fe0f)\b/g, ''))

const emojiNames = _.filter(_.map(index, (code, name) => code && _.escapeRegExp(name)))
const namesRe = new RegExp(':(' + emojiNames.join('|') + '):', 'g')

export function nameToEmojiID(name) {
  return index[name] || null
}

export function emojiIDToName(code) {
  return revIndex[code] || null
}

export function iconClass(emojiID) {
  return 'emoji emoji-' + emojiID.replace(/^~|-(200d|fe0f)\b/g, '')
}

export function nameToUnicode(name) {
  // Deliberately using the "full" index to avoid converting, e.g., :+1: to :thumbsup:.
  const code = index[name]
  if (!code || /^~/.test(code)) {
    return null
  }
  return code.split('-').map((cp) => String.fromCodePoint(Number.parseInt(cp, 16))).join('')
}

export function unicodeToEmojiID(string) {
  /* eslint-disable no-misleading-character-class */

  // If a U+FE0E VARIATION SELECTOR-14 (i.e. text-style display) appears anywhere, this
  // is probably not meant to be an emoji.
  if (/\ufe0e/.test(string)) return null

  // Normalize via the codepointIndex.
  // Despite its name, toCodePoint() handles multi-codepoint emoji correctly.
  return codepointIndex[twemoji.convert.toCodePoint(string.replace(/[\u200d\ufe0f]/g, ''))]
}

export default {
  unicodeIndex,
  index,
  revIndex,
  codepointIndex,
  namesRe,
  nameToEmojiID,
  emojiIDToName,
  iconClass,
  nameToUnicode,
  unicodeToEmojiID,
}
