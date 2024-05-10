if ('ontouchstart' in window) {
  document.body.classList.add('touch')
  document.body.addEventListener('touchstart', (ev) => {
    ev.target.classList.add('touching')
  }, false)
  document.body.addEventListener('touchend', (ev) => {
    ev.target.classList.remove('touching')
  }, false)
}
