(function initFastTouch(body) {
  if ('ontouchstart' in window) {
    body.classList.add('touch')
    body.addEventListener('touchstart', (ev) => {
      ev.target.classList.add('touching')
    }, false)
    body.addEventListener('touchend', (ev) => {
      ev.target.classList.remove('touching')
    }, false)
  }
}(document.body))
