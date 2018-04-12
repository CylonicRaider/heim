import React from 'react'
import PropTypes from 'prop-types'
import MarkdownIt from 'markdown-it'


const sectionRe = /^section (\w+)$/
const md = new MarkdownIt()
  .use(require('markdown-it-anchor'), {
    permalink: true,
    permalinkBefore: true,
    permalinkSymbol: '#',
  })
  .use(require('markdown-it-container'), 'section', {
    validate(params) {
      return params.trim().match(sectionRe)
    },

    render(tokens, idx) {
      const m = tokens[idx].info.trim().match(sectionRe)
      if (tokens[idx].nesting === 1) {
        return '<section class="' + m[1] + '">\n'
      }
      return '</section>\n'
    },
  })

export default function Markdown(props) {
  /* eslint-disable react/no-danger */
  return (
    <div className={props.className} dangerouslySetInnerHTML={{__html: md.render(props.content)}} />
  )
}

Markdown.propTypes = {
  className: PropTypes.string,
  content: PropTypes.string,
}
