import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'
import { TransitionGroup } from 'react-transition-group'
import Immutable from 'immutable'

import NotificationListItem from './NotificationListItem'
import Tree from '../Tree'


export default createReactClass({
  displayName: 'NotificationList',

  propTypes: {
    tree: PropTypes.instanceOf(Tree).isRequired,
    notifications: PropTypes.instanceOf(Immutable.OrderedMap).isRequired,
    onNotificationSelect: PropTypes.func,
    animate: PropTypes.bool,
  },

  render() {
    /* eslint-disable react/no-array-index-key */
    // let's hope this is actually intended...
    const notifications = this.props.notifications.map((kind, messageId) => <NotificationListItem key={messageId} tree={this.props.tree} nodeId={messageId} kind={kind} onClick={this.props.onNotificationSelect} />).toIndexedSeq()

    if (this.props.animate) {
      return (
        <TransitionGroup component="div" className="notification-list">
          {notifications}
        </TransitionGroup>
      )
    }
    return (
      <div className="notification-list">
        {notifications}
      </div>
    )
  },
})
