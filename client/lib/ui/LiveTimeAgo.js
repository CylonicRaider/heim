import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import Reflux from 'reflux'
import classNames from 'classnames'
import moment from 'moment'

import forwardProps from '../util/forwardProps'

function checkIsMoment(props, propName) {
  if (!moment.isMoment(props[propName])) {
    return new Error('not a Moment instance')
  }
  return null
}

export default createReactClass({
  displayName: 'LiveTimeAgo',

  propTypes: {
    time: PropTypes.oneOfType([PropTypes.number, checkIsMoment]),
    nowText: PropTypes.string,
    className: PropTypes.string,
  },

  mixins: [
    Reflux.connect(require('../stores/clock').minute, 'now'),
  ],

  render() {
    let t = this.props.time
    if (!moment.isMoment(t)) {
      t = moment.unix(t)
    }

    let display
    let className
    if (moment(this.state.now).diff(t, 'minutes') === 0) {
      display = this.props.nowText
      className = 'now'
    } else {
      display = t.locale('en-short').from(this.state.now, true)
    }

    return (
      <time dateTime={t.toISOString()} title={t.format('MMMM Do YYYY, h:mm:ss a')} {...forwardProps(this.props)} className={classNames(className, this.props.className)}>
        {display}
      </time>
    )
  },
})
