export default function findParent(el, predicate) {
  let curEl = el
  while (curEl) {
    if (predicate(curEl)) {
      return curEl
    }
    curEl = curEl.parentNode
  }
  return null
}
