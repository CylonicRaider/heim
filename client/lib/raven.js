import Raven from 'raven-js'
import consolePlugin from 'raven-js/plugins/console'

// Hack: Raven not currently setting window.Raven itself; perhaps this will be
// fixed in a future version.
window.Raven = Raven

Raven.config(process.env.SENTRY_ENDPOINT, {
  release: process.env.HEIM_RELEASE,
  tags: {git_commit: process.env.HEIM_GIT_COMMIT},
  autoBreadcrumbs: {console: false},
}).addPlugin(consolePlugin).install()

// Hack: After Raven is done with it, requestAnimationFrame stops working
// altogether (for me?)
window.requestAnimationFrame = function requestAnimationFrameShim(cb) {
  setTimeout(cb, 0)
}

const origCaptureException = Raven.captureException
window.Raven.captureException = function captureException(ex, options) {
  const newOptions = options || {}
  if (ex) {
    if (ex.action) {
      newOptions.tags = newOptions.tags || {}
      newOptions.tags.action = ex.action
    }
    if (ex.response) {
      newOptions.extra = newOptions.extra || {}
      newOptions.extra.response = ex.response
    }
  }
  origCaptureException.call(Raven, ex, newOptions)
}
