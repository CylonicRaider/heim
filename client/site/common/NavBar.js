import React from 'react'
import PropTypes from 'prop-types'
import classNames from 'classnames'

import heimURL from '../../lib/heimURL'

export default function NavBar(props) {
  return (
    <nav>
      <div className="container">
        <ul>
          <li className="home">
            <a href={heimURL('/')}>&nbsp;</a>
          </li>
          {props.items.map((item) => {
            let href = item.href
            if (href === undefined) href = item.name
            href = props.prefix + (href ? '/' : '') + href
            return (
              <li key={item.name} className={classNames(props.selected === item.name && 'selected')}>
                <a href={heimURL(href)}>{item.caption}</a>
              </li>
            )
          })}
        </ul>
      </div>
    </nav>
  )
}

NavBar.propTypes = {
  prefix: PropTypes.string,
  items: PropTypes.arrayOf(PropTypes.object),
  selected: PropTypes.string,
}
