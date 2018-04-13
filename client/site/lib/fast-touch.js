/* eslint-disable prefer-arrow-callback */
(function initFastTouch(body) {
  if ('ontouchstart' in window) {
    body.classList.add('touch')
    body.addEventListener('touchstart', function fastTouchStart(ev) {
      ev.target.classList.add('touching')
    }, false)
    body.addEventListener('touchend', function fastTouchEnd(ev) {
      ev.target.classList.remove('touching')
    }, false)
  }
}(document.body))
