import React from 'react'
import PropTypes from 'prop-types'
import classNames from 'classnames'

import Page from './Page'
import Footer from './Footer'

export default function MainPage(props) {
  return (
    <Page className={classNames('page', props.className)} title={props.title} heimPage={props.heimPage}>
      {props.nav || null}
      <div className="container main">
        {props.sidebar && <div className="sidebar-container">
          <div className="sidebar text-page" />
        </div>}
        {props.children}
      </div>
      <Footer />
    </Page>
  )
}

MainPage.propTypes = {
  className: PropTypes.string,
  title: PropTypes.string,
  heimPage: PropTypes.string,
  nav: PropTypes.node,
  sidebar: PropTypes.bool,
  children: PropTypes.node,
}
