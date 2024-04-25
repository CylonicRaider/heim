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
  .use((mdi) => {
    mdi.core.ruler.after('block', 'toc_anchor', (state) => {
      for (let i = 0; i < state.tokens.length; i++) {
        const token = state.tokens[i]
        if (token.type !== 'container_section_open') continue
        const m = token.info.trim().match(sectionRe)
        if (m[1] !== 'toc') continue

        const newToken = new state.Token('marker', 'a', 0)
        newToken.map = [token.map[0], token.map[0]]
        newToken.info = m[1]
        state.tokens.splice(i, 0, newToken)
        i++
      }
    })

    mdi.renderer.rules.marker = (tokens, idx) => (
      '<a id="marker-' + tokens[idx].info + '"></a>'
    )
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
