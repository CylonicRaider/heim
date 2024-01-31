/* eslint-disable import/prefer-default-export */

import ReactDOMServer from 'react-dom/server'

export function render(pageComponent) {
  return '<!DOCTYPE html>' + ReactDOMServer.renderToStaticMarkup(pageComponent)
}
